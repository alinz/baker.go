package engine

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/driver"
	"github.com/alinz/baker.go/internal/addr"
	"github.com/alinz/baker.go/internal/response"
)

func normalizeHost(host string) (string, bool) {
	www := false
	if strings.HasPrefix(host, "www.") {
		www = true
		host = strings.Replace(host, "www.", "", 1)
	}
	return host, www
}

type Engine struct {
	client  *http.Client
	store   *baker.Store
	watcher driver.Watcher
	mux     sync.RWMutex
}

var _ http.Handler = (*Engine)(nil)

func (e *Engine) configs(container *baker.Container) ([]*baker.Config, error) {
	configURL, err := addr.Join(container.RemoteAddr, container.ConfigPath)
	if err != nil {
		return nil, err
	}

	resp, err := e.client.Get(configURL)
	if err != nil {
		return nil, err
	}

	var configs []*baker.Config

	err = json.NewDecoder(resp.Body).Decode(&configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func (e *Engine) Start() error {
	// watcher
	watcher := make(chan *baker.Container, 100)
	go func() {
		defer close(watcher)

		for {
			container := e.watcher.Container()
			if container == nil {
				return
			}

			watcher <- container
		}
	}()

	// pulser
	pulser := make(chan *baker.Container, 100)
	go func() {
		defer close(pulser)

		deletedContainers := make([]*baker.Container, 0)
		containersMap := make(map[string]*baker.Container)
		pulse := time.After(10 * time.Second)

		for {
			select {
			case <-pulse:
				// send deleted items
				for _, container := range deletedContainers {
					pulser <- container
				}
				// clear deleted array
				deletedContainers = deletedContainers[:0]

				// loop over all items inside map
				for _, container := range containersMap {
					pulser <- container
				}

				// setup the next tick
				pulse = time.After(10 * time.Second)

			case container, ok := <-watcher:
				if !ok {
					return
				}

				if container.Active {
					containersMap[container.ID] = container
				} else {
					delete(containersMap, container.ID)
					deletedContainers = append(deletedContainers, container)
				}
			}
		}
	}()

	// pinger
	pinger := make(chan *baker.Target, 100)
	go func() {
		defer close(pinger)

		for container := range pulser {
			configs, err := e.configs(container)

			if err != nil {
				container.Err = err
				container.Active = false
				pinger <- &baker.Target{
					Container: container,
					Config:    nil,
				}
			} else {
				for _, config := range configs {
					pinger <- &baker.Target{
						Container: container,
						Config:    config,
					}
				}
			}
		}
	}()

	// updater
	for target := range pinger {
		e.mux.Lock()
		if !target.Container.Active || target.Container.Err != nil {
			e.store.Remove(target.Container)
		} else {
			e.store.Add(target.Container, target.Config)
		}
		e.mux.Unlock()
	}

	return nil
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	path := r.URL.Path

	e.mux.RLock()
	domain, err := e.store.Get(host)
	e.mux.RUnlock()
	if err != nil {
		response.AsJSON(w, http.StatusServiceUnavailable, err)
		return
	}

	e.mux.RLock()
	service, err := domain.Get(path)
	e.mux.RUnlock()
	if err != nil {
		response.AsJSON(w, http.StatusServiceUnavailable, err)
		return
	}

	e.mux.RLock()
	target, err := service.Get()
	e.mux.RUnlock()
	if err != nil {
		response.AsJSON(w, http.StatusServiceUnavailable, err)
		return
	}

	remoteAddr, err := url.Parse(addr.RemoteHTTP(target.Container.RemoteAddr, target.Container.ConfigPath, false).String())
	if err != nil {
		response.AsJSON(w, http.StatusServiceUnavailable, err)
		return
	}

	var handler http.Handler = httputil.NewSingleHostReverseProxy(remoteAddr)
	for i := len(target.Config.RuleHandlers); i > 0; i-- {
		handler = target.Config.RuleHandlers[i-1].ApplyRule(handler)
	}

	r.URL.Path = ""

	handler.ServeHTTP(w, r)
}

// New creates a new Reverse Proxy Engine based on given driver
func New(watcher driver.Watcher) *Engine {
	return &Engine{
		store:   baker.NewStore(),
		watcher: watcher,
		client:  &http.Client{},
	}
}
