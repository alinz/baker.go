package baker_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/alinz/baker.go"
)

type MockedWatcher struct {
	containers chan *baker.Container
}

var _ baker.Watcher = (*MockedWatcher)(nil)

func (w *MockedWatcher) Container() (*baker.Container, error) {
	container, ok := <-w.containers
	if !ok {
		return nil, baker.ErrWatcherClosed
	}
	return container, nil
}

func (w *MockedWatcher) push(container *baker.Container) {
	w.containers <- container
}

func (w *MockedWatcher) close() {
	close(w.containers)
}

func NewMockedWatcher() *MockedWatcher {
	return &MockedWatcher{
		containers: make(chan *baker.Container, 10),
	}
}

func createActiveMockedContainer(id, domain, path string) *baker.Container {

	container := &baker.Container{
		ID:     id,
		Active: true,
		RemoteAddr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: int(0),
		},
	}

	return container
}

func TestBasePager(t *testing.T) {
	watcher := NewMockedWatcher()
	pinger := baker.NewBasePinger(watcher)

	go func() {
		for {
			service, err := pinger.Service()
			if err != nil {
				return
			}

			fmt.Println(service)
		}
	}()

}
