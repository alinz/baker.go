package docker

import (
	"encoding/json"

	"github.com/alinz/baker.go"
)

type Container struct {
	client    Client
	id        string
	addr      string
	configURL string
}

var _ baker.Container = (*Container)(nil)
var _ baker.EndpointsFetcher = (*Container)(nil)

func (c *Container) ID() string {
	return c.id
}

func (c *Container) Addr() string {
	return c.addr
}

func (c *Container) FetchEndpoints() ([]*baker.Endpoint, error) {
	r, err := c.client.Get(c.configURL)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var endpoints []*baker.Endpoint
	err = json.NewDecoder(r).Decode(&endpoints)
	if err != nil {
		return nil, err
	}

	return endpoints, nil
}
