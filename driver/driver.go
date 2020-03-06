package driver

import (
	"context"

	"github.com/alinz/baker.go"
)

type Watcher interface {
	// Containers gets the channel of cotnainer object
	// context can be used to close watcher
	Containers(ctx context.Context) <-chan *baker.Container
}
