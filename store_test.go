package baker_test

import (
	"testing"
	"time"

	"github.com/alinz/baker.go"
)

type MockedPinger struct {
	services chan *baker.Service
}

func (m *MockedPinger) push(service *baker.Service) {
	m.services <- service
}

func (m *MockedPinger) close() {
	close(m.services)
}

func (m *MockedPinger) Service() (*baker.Service, error) {
	service, ok := <-m.services
	if !ok {
		return nil, baker.ErrPingerClosed
	}
	return service, nil
}

func NewMockedPinger() *MockedPinger {
	return &MockedPinger{
		services: make(chan *baker.Service, 10),
	}
}

func TestBaseStore(t *testing.T) {
	pinger := NewMockedPinger()
	store := baker.NewBaseStore(pinger)

	pinger.push(&baker.Service{
		Rule: &baker.Rule{
			Domain: "example.com",
			Path:   "/sample",
			Ready:  true,
		},
		Container: &baker.Container{
			ID:     "1",
			Active: true,
		},
	})

	time.Sleep(1 * time.Second)

	service := store.Query("example.com", "/sample")
	if service == nil {
		t.Fatal("expect to get at least one service but got nothing")
	}

	if service.Container.ID != "1" {
		t.Fatalf("expect to get service with container id 1, but got %s", service.Container.ID)
	}

	pinger.push(&baker.Service{
		Rule: &baker.Rule{
			Domain: "example.com",
			Path:   "/sample",
			Ready:  true,
		},
		Container: &baker.Container{
			ID:     "1",
			Active: false,
		},
	})

	time.Sleep(1 * time.Second)

	service = store.Query("example.com", "/sample")
	if service != nil {
		t.Fatal("expect receiving no service but got one")
	}
}
