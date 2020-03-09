package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/driver"
	"github.com/alinz/baker.go/internal/addr"
)

func join(paths ...string) (string, error) {
	u, err := url.Parse(paths[0])
	if err != nil {
		return "", err
	}

	paths[0] = u.Path

	u.Path = path.Join(paths...)
	return u.String(), nil
}

// DefaultClient is a default client which uses unix protocol
var DefaultClient = &http.Client{
	Transport: &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", "/var/run/docker.sock")
		},
	},
}

// DefaultAddr is a default docker host and port which communicating with Docker deamon
const DefaultAddr = "http://localhost"

type Event struct {
	ID     string `json:"id"`
	Active bool   `json:"active"`
}

type Watcher struct {
	client *http.Client
	addr   string
	events <-chan *Event
}

var _ driver.Watcher = (*Watcher)(nil)

func (w *Watcher) Container() *baker.Container {
	event, ok := <-w.events
	if !ok {
		return nil
	}

	return w.load(event)
}

func (w *Watcher) load(event *Event) *baker.Container {
	url, err := join(w.addr, "/containers/", event.ID, "/json")
	if err != nil {
		return &baker.Container{
			ID:  event.ID,
			Err: err,
		}
	}

	resp, err := w.client.Get(url)
	if err != nil {
		return &baker.Container{
			ID:  event.ID,
			Err: err,
		}
	}

	payload := &struct {
		ID string `json:"Id"`

		Config *struct {
			Labels *struct {
				Network     string `json:"baker.network"`
				ServicePort string `json:"baker.service.port"`
				ServicePing string `json:"baker.service.ping"`
				ServiceSSL  string `json:"baker.service.ssl"`
			} `json:"Labels"`
		} `json:"Config"`

		NetworkSettings struct {
			Networks map[string]struct {
				IPAddress string `json:"IPAddress"`
			} `json:"Networks"`
		} `json:"NetworkSettings"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(payload)
	if err != nil {
		return &baker.Container{
			ID:  event.ID,
			Err: err,
		}
	}

	network, ok := payload.NetworkSettings.Networks[payload.Config.Labels.Network]
	if !ok {
		return &baker.Container{
			ID:  event.ID,
			Err: fmt.Errorf("network '%s' not exists in labels", payload.Config.Labels.Network),
		}
	}

	port, err := strconv.ParseInt(payload.Config.Labels.ServicePort, 10, 32)
	if err != nil {
		return &baker.Container{
			ID:  event.ID,
			Err: fmt.Errorf("failed to parse port for container '%s' because %s", event.ID, err),
		}
	}

	return &baker.Container{
		ID:         event.ID,
		Active:     true,
		ConfigPath: payload.Config.Labels.ServicePing,
		RemoteAddr: addr.Remote(network.IPAddress, int(port)),
	}
}

func (w *Watcher) eventsChannel(size int) (<-chan *Event, error) {
	url, err := join(w.addr, "/containers/json")
	if err != nil {
		return nil, err
	}

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, err
	}

	payload := []struct {
		ID    string `json:"Id"`
		State string `json:"State"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}

	events := make(chan *Event, size)

	url, err = join(w.addr, "/events")
	if err != nil {
		return nil, err
	}

	resp, err = w.client.Get(url)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)

	go func() {
		defer close(events)

		for _, event := range payload {
			events <- &Event{
				ID:     event.ID,
				Active: event.State == "running",
			}
		}

		for {
			payload := struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			}{}

			err := decoder.Decode(&payload)
			if err != nil {
				break
			}

			if payload.Status != "die" && payload.Status != "start" {
				continue
			}

			events <- &Event{
				ID:     payload.ID,
				Active: payload.Status == "start",
			}
		}
	}()

	return events, nil
}

func (w *Watcher) Start() error {
	events, err := w.eventsChannel(10)
	if err != nil {
		return err
	}

	w.events = events
	return nil
}

func New(client *http.Client, addr string) *Watcher {
	return &Watcher{
		client: client,
		addr:   addr,
	}
}
