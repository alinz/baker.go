package balance

import (
	"sync"
)

type RoundRobin struct {
	keys []string
	mtx  sync.Mutex
	i    int
}

var _ Balancer = (*RoundRobin)(nil)

func (rb *RoundRobin) Add(key string) error {
	rb.mtx.Lock()
	defer rb.mtx.Unlock()

	for _, innerKey := range rb.keys {
		if innerKey == key {
			return ErrDuplicateKey
		}
	}

	rb.keys = append(rb.keys, key)

	return nil
}

func (rb *RoundRobin) Del(key string) error {
	rb.mtx.Lock()
	defer rb.mtx.Unlock()

	for i, innerKey := range rb.keys {
		if innerKey == key {
			rb.keys = append(rb.keys[:i], rb.keys[i+1:]...)
			return nil
		}
	}

	return ErrNotFound
}

func (rb *RoundRobin) Get() (string, error) {
	rb.mtx.Lock()
	defer rb.mtx.Unlock()

	n := len(rb.keys)

	if n == 0 {
		return "", ErrEmpty
	}

	rb.i++
	if rb.i >= n {
		rb.i = 0
	}

	return rb.keys[rb.i], nil
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		keys: make([]string, 0),
	}
}
