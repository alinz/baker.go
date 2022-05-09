package collection

import "sync"

type Map[T any] struct {
	rw         sync.RWMutex
	collection map[string]T
}

func (m *Map[T]) Put(key string, val T) {
	m.rw.Lock()
	defer m.rw.Unlock()

	m.collection[key] = val
}

func (m *Map[T]) Get(key string) (val T, ok bool) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	val, ok = m.collection[key]
	return
}

func (s *Map[T]) Len() int {
	s.rw.RLock()
	defer s.rw.RUnlock()

	return len(s.collection)
}

func (m *Map[T]) Delete(key string) {
	m.rw.Lock()
	defer m.rw.Unlock()

	delete(m.collection, key)
}

func NewMap[T any]() *Map[T] {
	return &Map[T]{
		collection: make(map[string]T),
	}
}
