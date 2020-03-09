package docker

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"path"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/driver"
)

func join(host string, p string) (string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, p)
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

type Watcher struct {
	client *http.Client
	addr   string
}

var _ driver.Watcher = (*Watcher)(nil)

func (w *Watcher) Container() *baker.Container {
	return nil
}

func (w *Watcher) Events(size int) (<-chan *Event, error) {
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

func New(client *http.Client, addr string) *Watcher {
	return &Watcher{
		client: client,
		addr:   addr,
	}
}
