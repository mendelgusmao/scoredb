package set

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"log"
)

type Set[V any] struct {
	items map[uint64]V
}

func New[V any](items ...V) *Set[V] {
	set := &Set[V]{items: make(map[uint64]V)}

	for _, item := range items {
		set.Insert(item)
	}

	return set
}

func (s *Set[V]) Insert(item V) {
	hash := s.hash(item)

	if oldItem, ok := s.items[hash]; ok {
		serializedItem := fmt.Sprintf("%v", item)
		serializedOldItem := fmt.Sprintf("%v", oldItem)

		if serializedItem != serializedOldItem {
			log.Printf("set.Insert: possible collision: %d (`%v` / `%v`)", hash, item, oldItem)
		}
	}

	s.items[hash] = item
}

func (s *Set[V]) Len() int {
	return len(s.items)
}

func (s *Set[V]) Do(f func(V)) {
	for _, item := range s.items {
		f(item)
	}
}

func (s *Set[V]) hash(item V) uint64 {
	h := fnv.New64()
	h.Write([]byte(fmt.Sprintf("%v", item)))

	return h.Sum64()
}

func (s *Set[V]) GobEncode() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buffer)

	if err := enc.Encode(s.items); err != nil {
		return nil, fmt.Errorf("[Set] %v", err)
	}

	return buffer.Bytes(), nil
}

func (s *Set[V]) GobDecode(input []byte) error {
	buffer := bytes.NewBuffer(input)
	dec := gob.NewDecoder(buffer)

	if err := dec.Decode(&s.items); err != nil {
		return fmt.Errorf("[Set] %v", err)
	}

	return nil
}
