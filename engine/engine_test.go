package engine_test

import (
	"context"
	"testing"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/driver"
	"github.com/alinz/baker.go/engine"
)

type mockedWatcher struct {
	canel      func()
	containers chan *baker.Container
}

func (mw *mockedWatcher) Containers(ctx context.Context) <-chan *baker.Container {
	return mw.containers
}

func mockWatcher(containers ...*baker.Container) driver.Watcher {
	chanContainers := make(chan *baker.Container, 1)

	go func() {
		defer close(chanContainers)
		for _, container := range containers {
			chanContainers <- container
		}
	}()

	return &mockedWatcher{
		containers: chanContainers,
	}
}

func TestEngine(t *testing.T) {

	engine.New()
}
