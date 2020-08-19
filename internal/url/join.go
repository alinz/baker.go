package url

import (
	"net"
	"net/url"
	"path"
	"strings"
)

// Join joins proper segments of URL. The first argument should be a URL
func Join(base net.Addr, paths ...string) (string, error) {
	builder := strings.Builder{}

	builder.WriteString("http://")
	builder.WriteString(base.String())

	u, err := url.Parse(builder.String())
	if err != nil {
		return "", err
	}

	u.Path = path.Join(paths...)
	return u.String(), nil
}
