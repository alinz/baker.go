package engine_test

import (
	"testing"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/driver"
	"github.com/alinz/baker.go/engine"
	"github.com/alinz/baker.go/internal/rule"
)

type mockedWatcher struct {
	containers chan *baker.Container
}

func (mw *mockedWatcher) Container() *baker.Container {
	return <-mw.containers
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
	err := rule.Register(&rule.PathReplaceRegistry{})
	if err != nil {
		t.Fatal(err)
	}

	watcher := mockWatcher()

	proxy := engine.New(watcher)

	err = proxy.Start()
	if err != nil {
		t.Fatal(err)
	}

}
