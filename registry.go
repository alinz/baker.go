package baker

import (
	"sync"
)

type Registry struct {
	domains map[string]*Services
	mux     sync.RWMutex
}

var _ ContainerRegistor = (*Registry)(nil)

func (r *Registry) UpdateContainer(container Container, endpoint *Endpoint) {
	r.mux.Lock()
	defer r.mux.Unlock()

	services, ok := r.domains[endpoint.Domain]
	if !ok && !endpoint.Ready {
		return
	} else if !ok {
		services = NewServices()
		r.domains[endpoint.Domain] = services
	}

	if endpoint.Ready {
		services.Add(endpoint, container)
	} else {
		services.Remove(endpoint, container)
	}
}

func (r *Registry) FindContainer(domain string, path string) (Container, *Endpoint) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	services, ok := r.domains[domain]
	if !ok {
		return nil, nil
	}

	return services.Get(path)
}

func NewRegistry() *Registry {
	return &Registry{
		domains: make(map[string]*Services),
	}
}
