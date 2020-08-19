package baker

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/alinz/baker.go/internal/logger"
	"github.com/alinz/baker.go/internal/store"
	"github.com/alinz/baker.go/internal/store/trie"
)

type BaseStore struct {
	closed  chan struct{}
	domains *Domains
}

var _ Store = (*BaseStore)(nil)

func (br *BaseStore) Query(domain, path string) *Service {
	return br.domains.Get(domain, path)
}

// NewBaseStore creates BaseStore object
func NewBaseStore(pinger Pinger) *BaseStore {
	router := &BaseStore{
		closed:  make(chan struct{}, 1),
		domains: NewDomains(),
	}

	go func() {
		defer func() {
			close(router.closed)
		}()

		for {
			service, err := pinger.Service()
			if errors.Is(err, ErrPingerClosed) {
				break
			}

			if !service.Container.Active || !service.Rule.Ready || service.Container.Err != nil {
				logger.Debug("STORE: removing container %s", service.Container.ID)
				router.domains.Remove(service)
			} else {
				logger.Debug("STORE: adding a new container %s", service.Container.ID)
				router.domains.Add(service)
			}
		}
	}()

	return router
}

//////////////////////////////

type Domains struct {
	collection map[string]*Paths
	mux        sync.RWMutex
}

func (d *Domains) Add(service *Service) {
	d.mux.Lock()
	defer d.mux.Unlock()

	path, ok := d.collection[service.Rule.Domain]
	if !ok {
		path = NewPaths()
		d.collection[service.Rule.Domain] = path
	}

	path.Add(service)
}

func (d *Domains) Remove(service *Service) {
	d.mux.Lock()
	defer d.mux.Unlock()

	path, ok := d.collection[service.Rule.Domain]
	if !ok {
		return
	}

	// TODO: do we need to remove rule.Domain value if there is no more
	// container available for that domain?

	path.Remove(service)
}

func (d *Domains) Get(domain, path string) *Service {
	d.mux.RLock()
	defer d.mux.RUnlock()

	paths, ok := d.collection[domain]
	if !ok {
		return nil
	}

	endpoints := paths.Get(path)
	if endpoints == nil {
		return nil
	}

	return endpoints.Get()
}

func NewDomains() *Domains {
	return &Domains{
		collection: make(map[string]*Paths),
	}
}

type Paths struct {
	collection store.KeyValue
	mux        sync.RWMutex
}

func (p *Paths) Add(service *Service) {
	p.mux.Lock()
	defer p.mux.Unlock()

	key := []rune(service.Rule.Path)

	var endpoints *Endpoints

	value, err := p.collection.Get(key)
	if err != nil {
		endpoints = NewEndpoints()
		p.collection.Put(key, endpoints)
	} else {
		endpoints = value.(*Endpoints)
	}

	endpoints.Add(service)
}

func (p *Paths) Remove(service *Service) {
	p.mux.Lock()
	defer p.mux.Unlock()

	key := []rune(service.Rule.Path)

	value, err := p.collection.Get(key)
	if err != nil {
		return
	}

	endpoints := value.(*Endpoints)
	endpoints.Remove(service)

	//p.collection.Del(key)
}

func (p *Paths) Get(path string) *Endpoints {
	p.mux.RLock()
	defer p.mux.RUnlock()

	key := []rune(path)

	value, err := p.collection.Get(key)
	if err != nil {
		return nil
	}

	return value.(*Endpoints)
}

func NewPaths() *Paths {
	return &Paths{
		collection: trie.New(true),
	}
}

type Endpoints struct {
	collection []*Service
	next       uint32
	mux        sync.RWMutex
}

func (e *Endpoints) Add(service *Service) {
	e.mux.Lock()
	defer e.mux.Unlock()

	for i, s := range e.collection {
		if s.Container.ID == service.Container.ID {
			e.collection[i] = service
			return
		}
	}

	e.collection = append(e.collection, service)
}

func (s *Endpoints) Remove(service *Service) {
	s.mux.Lock()
	defer s.mux.Unlock()

	id := service.Container.ID

	for i, service := range s.collection {
		if service.Container.ID == id {
			logger.Debug("STORE: remove container %s from endpoint list", id)
			copy(s.collection[i:], s.collection[i+1:])
			s.collection[len(s.collection)-1] = nil // This is for make sure GC is be able to clean object
			s.collection = s.collection[:len(s.collection)-1]
			break
		}
	}
}

func (s *Endpoints) Get() *Service {
	s.mux.RLock()
	defer s.mux.RUnlock()

	size := len(s.collection)
	if size == 0 {
		return nil
	}

	n := atomic.AddUint32(&s.next, 1)

	return s.collection[(int(n)-1)%size]
}

func NewEndpoints() *Endpoints {
	return &Endpoints{
		collection: make([]*Service, 0),
	}
}
