package baker

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/alinz/baker.go/internal/logger"
	"github.com/alinz/baker.go/internal/timer"
	"github.com/alinz/baker.go/internal/url"
)

type BasePinger struct {
	closed           chan struct{}
	containersMap    map[string]*Container
	containersMapMux sync.Mutex
	containers       chan *Container
	servicesMap      map[string]*Services // to keep old service information
	services         chan *Service
	cancelPumps      func()
	client           *http.Client
}

var _ Pinger = (*BasePinger)(nil)

// Service calls each available container and create a service object
func (p *BasePinger) Service() (*Service, error) {
	select {
	case <-p.closed:
		return nil, ErrPingerClosed
	case service, ok := <-p.services:
		if !ok {
			return nil, ErrPingerClosed
		}
		return service, nil
	}
}

// pings retrive each container from containers channel and attach
// and creates one or more services
func (p *BasePinger) pings() {
	defer func() {
		close(p.services)
	}()

	for {
		var container *Container
		var ok bool

		select {
		case <-p.closed:
			return
		case container, ok = <-p.containers:
			if !ok {
				logger.Error("PINGER: failed to fetch container")
				continue
			}
		}

		configURL, err := url.Join(container.RemoteAddr, container.ConfigPath)
		if err != nil {
			logger.Warn("PINGER: failed to construct config url: %s", err)
			continue
		}

		services, ok := p.servicesMap[container.ID]
		if !ok {
			services = NewServices()
			p.servicesMap[container.ID] = services
		}

		resp, err := p.client.Get(configURL)
		if err != nil {
			logger.Warn("PINGER: failed to ping %s to retrive service rule: %s", configURL, err)

			if ok {
				services.Each(func(service *Service, idx int) {
					service.Container.Active = false
					select {
					case <-time.After(1 * time.Second):
						logger.Warn("PINGER: too slow to pump service %s to services channel", container.ID)
					case p.services <- service:
						// ignore
					}
				})
			} else {
				p.containersMapMux.Lock()
				delete(p.containersMap, container.ID)
				p.containersMapMux.Unlock()
			}

			continue
		}

		var rules Rules
		err = json.NewDecoder(resp.Body).Decode(&rules)
		if err != nil {
			logger.Warn("failed to parse body received from %s for service rule: %s", configURL, err)
			continue
		}

		services.Clean()

		rules.Each(func(rule *Rule) {
			service := &Service{
				Container: container,
				Rule:      rule,
			}

			services.Add(service)

			select {
			case <-time.After(1 * time.Second):
				logger.Warn("PINGER: too slow to pump service %s to services channel", container.ID)
			case p.services <- service:
				// ignore
			}
		})
	}
}

// pumps perodically scans containersMap and push each container to containers channel
func (p *BasePinger) pumps() {
	if p.cancelPumps != nil {
		p.cancelPumps()
	}

	p.cancelPumps = timer.Interval(10*time.Second, func(ctx context.Context) {
		p.containersMapMux.Lock()
		defer p.containersMapMux.Unlock()

		deletedContainers := make([]string, 0)

		for id, container := range p.containersMap {
			// add containers to deletedContainer list that are no longer needs to be processed
			if !container.Active || container.Err != nil {
				deletedContainers = append(deletedContainers, id)
			}

			// this select prevents slow push to containers channel
			select {
			case <-time.After(1 * time.Second):
				logger.Warn("PINGER: too slow to pump container %s to containers channel", id)
			case p.containers <- container:
			}
		}

		// delete the containers that no longer needed from the map
		for _, id := range deletedContainers {
			delete(p.containersMap, id)
		}
	})
}

func NewBasePinger(watcher Watcher) *BasePinger {
	pinger := &BasePinger{
		closed:        make(chan struct{}, 1),
		containersMap: make(map[string]*Container),
		containers:    make(chan *Container, 10),
		servicesMap:   make(map[string]*Services),
		services:      make(chan *Service, 10),
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 2 * time.Second,
				}).Dial,
			},
		},
	}

	go pinger.pings()
	go pinger.pumps()

	go func() {
		defer func() {
			pinger.cancelPumps()
			close(pinger.closed)
		}()

		fetch := func() bool {
			container, err := watcher.Container()

			if errors.Is(err, ErrWatcherClosed) {
				// watcher has been closed, so pinger needs to be closed as well
				return true
			} else if err != nil {
				logger.Warn("PINGER: failed to fetch container from watcher: %s", err)
				return false
			}

			pinger.containersMapMux.Lock()
			defer pinger.containersMapMux.Unlock()

			// if we've seen this container before, update error and active properties
			if foundContainer, ok := pinger.containersMap[container.ID]; ok {
				foundContainer.Active = container.Active
				foundContainer.Err = container.Err

				logger.Debug("PINGER: updated container %s", container.ID)

				return false
			}

			// this is a new container, but has some error or being shutdown,
			// so ignore this container
			if container.Err != nil || !container.Active {
				// ignore this container
				logger.Debug("PINGER: ignored container %s", container.ID)
				return false
			}

			// found a new container that hasn't been seen before
			// adding it to the map
			pinger.containersMap[container.ID] = container
			logger.Debug("PINGER: added a new container %s", container.ID)

			return false
		}

		for {
			if done := fetch(); done {
				break
			}
		}
	}()

	return pinger
}
