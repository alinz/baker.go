package driver

import (
	"github.com/alinz/baker.go"
)

// Watcher defines how driver should react
type Watcher interface {
	// Container gets the next availabel container
	// this is a blocking calls
	Container() *baker.Container
}
