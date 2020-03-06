package docker

import (
	"context"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/driver"
)

type Watcher struct{}

var _ driver.Watcher = (*Watcher)(nil)

func (w *Watcher) Containers(ctx context.Context) <-chan *baker.Container {
	return nil
}

func NewWaWatcher() *Watcher {
	return &Watcher{}
}
