package driver

import (
	"github.com/alinz/baker.go"
)

type Watcher interface {
	// Container gets the next availabel container
	Container() *baker.Container
}
