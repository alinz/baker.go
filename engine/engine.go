package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/driver"
	"github.com/alinz/baker.go/internal/acme"
	"github.com/alinz/baker.go/internal/addr"
	"github.com/alinz/baker.go/internal/errors"
	"github.com/alinz/baker.go/internal/response"
	"github.com/alinz/baker.go/pkg/logger"
)

var log = logger.Default

var ErrBadRemoteAddr = errors.Value("bad remote address")
var ErrServiceNotAvailable = errors.Value("service not available")

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
var _ acme.PolicyManager = (*Engine)(nil)

func (e *Engine) configs(container *baker.Container) ([]*baker.Config, error) {
	if container.RemoteAddr == nil {
		return nil, ErrBadRemoteAddr
	}

	configURL, err := addr.Join(container.RemoteAddr, container.ConfigPath)
	if err != nil {
		return nil, err
	}

	resp, err := e.client.Get(configURL)
	if err != nil {
		log.Error("failed to get config from %s because %s", configURL, err)
		return nil, err
	}

	log.Debug("request config from %s has status code %d", configURL, resp.StatusCode)

	var configs []*baker.Config

	err = json.NewDecoder(resp.Body).Decode(&configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func (e *Engine) Start() {
	// watcher
	watcher := make(chan *baker.Container, 100)
	go func() {
		defer func() {
			log.Debug("watcher pipe is closed")
			close(watcher)
		}()

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
		defer func() {
			log.Debug("pulser pipe is closed")
			close(pulser)
		}()

		containersMap := make(map[string]*baker.Container)
		pulse := time.After(10 * time.Second)

		for {
			select {
			case container, ok := <-watcher:
				if !ok {
					return
				}

				if container.Active {
					containersMap[container.ID] = container
				} else {
					delete(containersMap, container.ID)
				}

				pulser <- container

			case <-pulse:
				log.Debug("updated %d containers", len(containersMap))

				// loop over all items inside map
				for _, container := range containersMap {
					pulser <- container
				}

				// setup the next tick
				pulse = time.After(10 * time.Second)
			}
		}
	}()

	// pinger
	pinger := make(chan *baker.Target, 100)
	go func() {
		defer func() {
			log.Debug("pinger pipe is closed")
			close(pinger)
		}()

		for container := range pulser {
			if !container.Active {
				pinger <- &baker.Target{
					Container: container,
					Config:    nil,
				}
				continue
			}

			configs, err := e.configs(container)

			if err != nil {
				container.Err = err
				container.Active = false
				pinger <- &baker.Target{
					Container: container,
					Config:    nil,
				}

				log.Error("Failed to ping %s container: %s", container.ID, err)
				continue
			}

			for _, config := range configs {
				pinger <- &baker.Target{
					Container: container,
					Config:    config,
				}
			}
		}
	}()

	// updater
	go func() {
		defer func() {
			log.Debug("updater pipe is closed")
		}()

		for target := range pinger {
			e.mux.Lock()
			if !target.Container.Active || target.Container.Err != nil {
				e.store.Remove(target.Container)
			} else {
				e.store.Add(target.Container, target.Config)
			}
			e.mux.Unlock()
		}
	}()
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	path := r.URL.Path

	e.mux.RLock()
	domain, err := e.store.Get(host)
	e.mux.RUnlock()
	if err != nil {
		log.Warn("failed to get any domain for %s and %s becuase of %s", host, path, err)
		response.AsJSON(w, http.StatusServiceUnavailable, ErrServiceNotAvailable)
		return
	}

	e.mux.RLock()
	service, err := domain.Get(path)
	e.mux.RUnlock()
	if err != nil {
		log.Warn("failed to get any service for %s and %s becuase of %s", host, path, err)
		response.AsJSON(w, http.StatusServiceUnavailable, ErrServiceNotAvailable)
		return
	}

	e.mux.RLock()
	target, err := service.Get()
	e.mux.RUnlock()
	if err != nil {
		log.Warn("failed to get any target for %s and %s becuase of %s", host, path, err)
		response.AsJSON(w, http.StatusServiceUnavailable, ErrServiceNotAvailable)
		return
	}

	remoteAddr, err := url.Parse(addr.RemoteHTTP(target.Container.RemoteAddr, path, false).String())
	if err != nil {
		log.Warn("failed to create remote address for %s and %s because of %s", target.Container.RemoteAddr, path, err)
		response.AsJSON(w, http.StatusServiceUnavailable, ErrServiceNotAvailable)
		return
	}

	if !target.Config.Ready {
		response.AsJSON(w, http.StatusServiceUnavailable, ErrServiceNotAvailable)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remoteAddr)
	director := proxy.Director
	proxy.Director = func(r *http.Request) {
		// Need to clear URL.Path to empty as target is already known
		// Also, NewSingleHostReverseProxy.Director's default
		// will try to merge target.Path and r.URL.Path
		r.URL.Path = ""
		director(r)
		// for some reasons, original director inside NewSingleHostReverseProxy add extra /
		// there is no point to have that so in this section, we are removing it
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")

		for i := 0; i < len(target.Config.RuleHandlers); i++ {
			target.Config.RuleHandlers[i].ApplyRule(r)
		}
	}

	proxy.ServeHTTP(w, r)
}

func (e *Engine) HostPolicy(ctx context.Context, host string) error {
	// www.example.com not store in domains.Paths
	// need to remove `www.` from it.
	h, _ := normalizeHost(host)

	e.mux.RLock()
	_, err := e.store.Get(h)
	e.mux.RUnlock()

	if err != nil {
		return err
	}

	return nil
}

// New creates a new Reverse Proxy Engine based on given driver
func New(watcher driver.Watcher) *Engine {
	return &Engine{
		store:   baker.NewStore(),
		watcher: watcher,
		client:  &http.Client{},
	}
}
