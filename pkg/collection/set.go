package collection

import (
	"math/rand"
	"sync"
)

type Set[K comparable, V any] struct {
	rw     sync.RWMutex
	mapper map[K]V
}

func (s *Set[K, V]) Put(key K, val V) {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.mapper[key] = val
}

func (s *Set[K, V]) Get(key K) (val V, ok bool) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	val, ok = s.mapper[key]
	return
}

func (s *Set[K, V]) Remove(key K) {
	s.rw.Lock()
	defer s.rw.Unlock()
	delete(s.mapper, key)
}

func (s *Set[K, V]) Contains(key K) bool {
	_, ok := s.Get(key)
	return ok
}

func (s *Set[K, V]) Iterate(fn func(key K, val V) bool) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	for key, val := range s.mapper {
		if !fn(key, val) {
			break
		}
	}
}

func (s *Set[K, V]) Random() (V, bool) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	var value V
	var keys []K

	if len(s.mapper) == 0 {
		return value, false
	}

	for key := range s.mapper {
		keys = append(keys, key)
	}

	value, ok := s.mapper[keys[rand.Intn(len(keys))]]
	return value, ok
}

func NewSet[K comparable, V any]() *Set[K, V] {
	return &Set[K, V]{
		mapper: make(map[K]V),
	}
}
