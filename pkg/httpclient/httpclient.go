package httpclient

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

type Getter interface {
	Get(url string) (io.ReadCloser, error)
}

type GetterFunc func(url string) (io.ReadCloser, error)

func (fn GetterFunc) Get(url string) (io.ReadCloser, error) {
	return fn(url)
}

func Unix(sockPath string) GetterFunc {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		},
	}

	return GetterFunc(func(url string) (io.ReadCloser, error) {
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	})
}

func New() GetterFunc {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	return GetterFunc(func(url string) (io.ReadCloser, error) {
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	})
}
