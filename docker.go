package baker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/alinz/baker.go/internal/logger"
)

type DockerWatcher struct {
	host       string
	client     *http.Client
	containers chan *Container
}

var _ Watcher = (*DockerWatcher)(nil)

func (d *DockerWatcher) load(id string) *Container {
	resource := d.host + "/containers/" + id + "/json"

	resp, err := d.client.Get(resource)
	if err != nil {
		return &Container{
			ID:  id,
			Err: err,
		}
	}

	payload := &struct {
		ID string `json:"Id"`

		Config *struct {
			Labels *struct {
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
	}{}

	err = json.NewDecoder(resp.Body).Decode(payload)
	if err != nil {
		return &Container{
			ID:  id,
			Err: err,
		}
	}

	if payload.Config.Labels.Enable != "true" {
		return &Container{
			ID:     id,
			Active: false,
		}
	}

	network, ok := payload.NetworkSettings.Networks[payload.Config.Labels.Network]
	if !ok {
		return &Container{
			ID:  id,
			Err: fmt.Errorf("network '%s' not exists in labels", payload.Config.Labels.Network),
		}
	}

	port, err := strconv.ParseInt(payload.Config.Labels.ServicePort, 10, 32)
	if err != nil {
		return &Container{
			ID:  id,
			Err: fmt.Errorf("failed to parse port for container '%s' because %s", id, err),
		}
	}

	return &Container{
		Type:       DockerContainer,
		ID:         id,
		Active:     true,
		ConfigPath: payload.Config.Labels.ServicePing,
		RemoteAddr: &net.TCPAddr{
			IP:   net.ParseIP(network.IPAddress),
			Port: int(port),
		},
	}
}

func (d *DockerWatcher) currentContainers() error {
	resource := d.host + "/containers/json"

	logger.Debug("WATCHER: getting list of current running containers")

	resp, err := d.client.Get(resource)
	if err != nil {
		return fmt.Errorf("failed to process http client: %w", err)
	}

	events := []struct {
		ID    string `json:"Id"`
		State string `json:"State"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&events)
	if err != nil {
		return fmt.Errorf("failed to decode events: %w", err)
	}

	logger.Debug("WATCHER: found %d running containers", len(events))

	for _, event := range events {
		var container *Container

		if event.State != "running" {
			container = &Container{
				ID:     event.ID,
				Active: false,
			}
			continue
		}

		container = d.load(event.ID)

		d.containers <- container
	}

	return nil
}

func (d *DockerWatcher) futureContainers() error {
	resource := d.host + "/events"

	logger.Debug("WATCHER: getting future containers")

	resp, err := d.client.Get(resource)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(resp.Body)

	event := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{}

	for {
		if err := decoder.Decode(&event); err != nil {
			return fmt.Errorf("failed to process event stream: %w", err)
		}

		logger.Debug("WATCHER: get a container %s with status %s", event.ID, event.Status)

		if event.Status != "die" && event.Status != "start" {
			continue
		}

		d.containers <- d.load(event.ID)
	}
}

// Start is a blocking call, the caller for this method
// should run in a separate go routine
//
// NOTE: upon on returning this function, the internal
// channel will be closed, if a restart requires, a new
// instance of DockerWatcher needs to be instantiated
func (d *DockerWatcher) Start() error {
	defer close(d.containers)

	err := d.currentContainers()
	if err != nil {
		return err
	}

	err = d.futureContainers()
	if err != nil {
		return err
	}

	return nil
}

func (d *DockerWatcher) Container() (*Container, error) {
	container, ok := <-d.containers
	if !ok {
		return nil, ErrWatcherClosed
	}

	return container, nil
}

type DockerWatcherConfig struct {
	Host   string
	Client *http.Client
	Size   int
}

var DefaultDockerWatcherConfig = DockerWatcherConfig{
	Host: "http://localhost",
	Client: &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	},
	Size: 10,
}

func NewDockerWatcher(config DockerWatcherConfig) *DockerWatcher {
	return &DockerWatcher{
		host:       config.Host,
		client:     config.Client,
		containers: make(chan *Container, config.Size),
	}
}
