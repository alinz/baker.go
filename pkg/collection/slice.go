package collection

import "sync"

type Slice[T any] struct {
	rw         sync.RWMutex
	collection []T
}

func (s *Slice[T]) Put(val T) {
	s.rw.Lock()
	defer s.rw.Unlock()

	s.collection = append(s.collection, val)
}

func (s *Slice[T]) Get(i int) (val T, ok bool) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	if i < len(s.collection) {
		val = s.collection[i]
		ok = true
	}

	return
}

func (s *Slice[T]) Delete(i int) {
	s.rw.Lock()
	defer s.rw.Unlock()

	if i < len(s.collection) {
		s.collection = append(s.collection[:i], s.collection[i+1:]...)
	}
}

func (s *Slice[T]) Len() int {
	s.rw.RLock()
	defer s.rw.RUnlock()

	return len(s.collection)
}

func (s *Slice[T]) Index(fn func(item T) bool) int {
	s.rw.RLock()
	defer s.rw.RUnlock()

	for i, item := range s.collection {
		if fn(item) {
			return i
		}
	}

	return -1
}

func NewSlice[T any]() *Slice[T] {
	return &Slice[T]{
		collection: make([]T, 0),
	}
}
