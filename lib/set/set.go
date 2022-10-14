package set

import (
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

func (s *Set[V]) hash(item any) uint64 {
	h := fnv.New64()
	h.Write([]byte(fmt.Sprintf("%v", item)))

	return h.Sum64()
}
