package docker

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

type Client interface {
	Get(url string) (io.ReadCloser, error)
}

type ClientFunc func(url string) (io.ReadCloser, error)

func (fn ClientFunc) Get(url string) (io.ReadCloser, error) {
	return fn(url)
}

func UnixClient() Client {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	return ClientFunc(func(url string) (io.ReadCloser, error) {
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	})
}

func RemoteClient() Client {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	return ClientFunc(func(url string) (io.ReadCloser, error) {
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	})
}
