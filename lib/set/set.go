package set

import (
	"fmt"
	"hash/fnv"
	"log"
)

type Set struct {
	items map[uint64]interface{}
}

func New(items ...interface{}) *Set {
	set := &Set{items: make(map[uint64]interface{})}

	for _, item := range items {
		set.Insert(item)
	}

	return set
}

func (s *Set) Insert(item interface{}) {
	hash := s.hash(item)

	if oldItem, ok := s.items[hash]; ok && item != oldItem {
		log.Printf("set.Insert: possible collision: %d (`%v` / `%v`)", hash, item, oldItem)
	}

	s.items[hash] = item
}

func (s *Set) Len() int {
	return len(s.items)
}

func (s *Set) Do(f func(interface{})) {
	for _, item := range s.items {
		f(item)
	}
}

func (s *Set) hash(item interface{}) uint64 {
	h := fnv.New64()
	h.Write([]byte(fmt.Sprintf("%v", item)))

	return h.Sum64()
}
