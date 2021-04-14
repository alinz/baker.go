package baker

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrWatcherClosed = errors.New("watcher is closed")
)

type Middleware interface {
	Process(next http.Handler) http.Handler
}

type Rule struct {
	Type string          `json:"type"`
	Args json.RawMessage `json:"args"`
}

type Endpoint struct {
	Domain string `json:"domain"`
	Path   string `json:"path"`
	Rules  []Rule `json:"rules"`
	Ready  bool   `json:"ready"`
}

type Container interface {
	ID() string
	Addr() string
}

type EndpointsFetcher interface {
	FetchEndpoints() ([]*Endpoint, error)
}

type ContainerRegistor interface {
	UpdateContainer(container Container, endpoint *Endpoint)
	FindContainer(domain, path string) (Container, *Endpoint)
}

type Watcher interface {
	Watch(errs chan<- error) <-chan Container
}
