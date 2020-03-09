package addr

import (
	"bytes"
	"fmt"
	"strings"
)

// Endpoint represents an address
type Endpoint interface {
	Host() string
	Port() int
	String() string
}

type remoteEndpoint struct {
	host string
	port int
}

var _ Endpoint = (*remoteEndpoint)(nil)

func (e remoteEndpoint) Host() string {
	return e.host
}

func (e remoteEndpoint) Port() int {
	return e.port
}

func (e remoteEndpoint) String() string {
	return fmt.Sprintf("%s:%d", e.host, e.port)
}

// Remote create and Endpoint
func Remote(host string, port int) Endpoint {
	return &remoteEndpoint{
		host: host,
		port: port,
	}
}

type httpRemoteEndpoint struct {
	Endpoint
	path   string
	secure bool
}

func (e *httpRemoteEndpoint) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("http")
	if e.secure {
		buffer.WriteString("s")
	}

	buffer.WriteString("://")
	buffer.WriteString(e.Endpoint.String())

	if !strings.HasPrefix(e.path, "/") {
		buffer.WriteString("/")
	}

	buffer.WriteString(e.path)

	return buffer.String()
}

// RemoteHTTP converts the plain address to http address
func RemoteHTTP(addr Endpoint, path string, secure bool) Endpoint {
	return &httpRemoteEndpoint{
		Endpoint: addr,
		path:     path,
		secure:   secure,
	}
}
