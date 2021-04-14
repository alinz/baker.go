package docker

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/alinz/baker.go"
)

type Watcher struct {
	unixClient   Client
	remoteClient Client
	closed       chan struct{}
}

var _ baker.Watcher = (*Watcher)(nil)

func (w *Watcher) load(id string) (*Container, error) {
	r, err := w.unixClient.Get("http://localhost/containers/" + id + "/json")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	payload := struct {
		Config struct {
			Labels struct {
				Enable      string `json:"baker.enable"`
				Network     string `json:"baker.network"`
				ServicePort string `json:"baker.service.port"`
				ServicePing string `json:"baker.service.ping"`
			} `json:"Labels"`
		} `json:"Config"`
		NetworkSettings struct {
			Networks map[string]struct {
				IPAddress string `json:"IPAddress"`
			} `json:"Networks"`
		} `json:"NetworkSettings"`
		ID string `json:"Id"`
	}{}

	err = json.NewDecoder(r).Decode(&payload)
	if err != nil {
		return nil, err
	}

	if payload.Config.Labels.Enable != "true" {
		return nil, fmt.Errorf("label 'baker.enable' is not set to true")
	}

	network, ok := payload.NetworkSettings.Networks[payload.Config.Labels.Network]
	if !ok {
		return nil, fmt.Errorf("network '%s' not exists in labels", payload.Config.Labels.Network)
	}

	port, err := strconv.ParseInt(payload.Config.Labels.ServicePort, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse port for container '%s' because %s", id, err)
	}

	addr := ""

	if network.IPAddress != "" {
		addr = fmt.Sprintf("%s:%d", network.IPAddress, port)
	}

	return &Container{
		id:        id,
		addr:      addr,
		client:    w.remoteClient,
		configURL: fmt.Sprintf("http://%s%s", addr, payload.Config.Labels.ServicePing),
	}, nil
}

func (w *Watcher) currentContainers(containers chan<- baker.Container, errs chan<- error) {
	r, err := w.unixClient.Get("http://localhost/containers/json")
	if err != nil {
		errs <- err
		return
	}
	defer r.Close()

	events := []struct {
		ID    string `json:"Id"`
		State string `json:"State"`
	}{}

	err = json.NewDecoder(r).Decode(&events)
	if err != nil {
		errs <- fmt.Errorf("failed to decode events: %w", err)
		return
	}

	for _, event := range events {
		var container *Container

		if event.State != "running" {
			continue
		}

		container, err := w.load(event.ID)
		if err != nil {
			errs <- fmt.Errorf("failed to load event %s: %w", event.ID, err)
			continue
		}

		select {
		case <-w.closed:
			return
		default:
			containers <- container
		}
	}
}

func (w *Watcher) futureContainers(containers chan<- baker.Container, errs chan<- error) {
	r, err := w.unixClient.Get("http://localhost/events")
	if err != nil {
		errs <- err
		return
	}
	defer r.Close()

	decoder := json.NewDecoder(r)

	event := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{}

	for {
		event.ID = ""
		event.Status = ""

		if err := decoder.Decode(&event); err != nil {
			errs <- fmt.Errorf("failed to decode json event stream: %w", err)
			continue
		}

		if event.Status != "die" && event.Status != "start" {
			continue
		}

		container, err := w.load(event.ID)
		if err != nil {
			errs <- fmt.Errorf("failed to load event %s: %w", event.ID, err)
			continue
		}

		select {
		case <-w.closed:
			return
		default:
			containers <- container
		}
	}
}

func (w *Watcher) Watch(errs chan<- error) <-chan baker.Container {
	containers := make(chan baker.Container, 10)

	w.closed = make(chan struct{}, 1)

	go func() {
		defer close(w.closed)

		w.currentContainers(containers, errs)
		w.futureContainers(containers, errs)
	}()

	return containers
}

func NewWatcher(client Client) *Watcher {
	watcher := &Watcher{
		unixClient:   client,
		remoteClient: RemoteClient(),
	}

	return watcher
}
