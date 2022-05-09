package docker

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"strconv"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/pkg/httpclient"
	"github.com/alinz/baker.go/pkg/log"
)

type Docker struct {
	unix   httpclient.GetterFunc
	http   httpclient.GetterFunc
	closed chan struct{}
}

func (d *Docker) loadByID(id string) (*baker.Container, error) {
	r, err := d.unix("http://localhost/containers/" + id + "/json")
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
		fmt.Println(payload.NetworkSettings.Networks)
		return nil, fmt.Errorf("network '%s' not exists in labels", payload.Config.Labels.Network)
	}

	port, err := strconv.ParseInt(payload.Config.Labels.ServicePort, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse port for container '%s' because %s", id, err)
	}

	var addr netip.AddrPort

	if network.IPAddress != "" {
		addr, err = netip.ParseAddrPort(fmt.Sprintf("%s:%d", network.IPAddress, port))
	}

	return &baker.Container{
		ID:   id,
		Addr: addr,
		Path: payload.Config.Labels.ServicePing,
	}, nil
}

func (d *Docker) currentContainers(containers chan<- *baker.Container) {
	r, err := d.unix("http://localhost/containers/json")
	if err != nil {
		log.Error().Err(err).Msg("failed to get containers")
		return
	}
	defer r.Close()

	events := []struct {
		ID    string `json:"Id"`
		State string `json:"State"`
	}{}

	err = json.NewDecoder(r).Decode(&events)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode containers")
		return
	}

	for _, event := range events {
		var container *baker.Container

		if event.State != "running" {
			continue
		}

		container, err := d.loadByID(event.ID)
		if err != nil {
			log.Debug().Err(err).Str("id", event.ID).Msg("Failed to load container")
			continue
		}

		select {
		case <-d.closed:
			return
		default:
			containers <- container
		}
	}
}

func (d *Docker) futureContainers(containers chan<- *baker.Container) {
	r, err := d.unix("http://localhost/events")
	if err != nil {
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
			log.Error().Err(err).Msg("failed to decode event")
			continue
		}

		if event.Status != "die" && event.Status != "start" {
			continue
		}

		container, err := d.loadByID(event.ID)
		if err != nil {
			log.Error().Err(err).Str("id", event.ID).Msg("Failed to load container")
			continue
		}

		select {
		case <-d.closed:
			return
		default:
			containers <- container
		}
	}
}

func New() (<-chan *baker.Container, error) {
	d := &Docker{
		unix:   httpclient.Unix("/var/run/docker.sock"),
		http:   httpclient.New(),
		closed: make(chan struct{}),
	}

	containers := make(chan *baker.Container, 10)

	go func() {
		defer close(d.closed)

		d.currentContainers(containers)
		d.futureContainers(containers)
	}()

	return containers, nil
}
