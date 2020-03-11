package main

import (
	"net/http"
	"os"

	"github.com/alinz/baker.go/driver/docker"
	"github.com/alinz/baker.go/engine"
	"github.com/alinz/baker.go/internal/acme"
	"github.com/alinz/baker.go/pkg/rule"
)

type server interface {
	Start(handler http.Handler) error
}

type serverFun func(handler http.Handler) error

func (s serverFun) Start(handler http.Handler) error {
	return s(handler)
}

func main() {
	acmePath := os.Getenv("BAKER_ACME_PATH")
	acmeEnable := os.Getenv("BAKER_ACME") == "YES"

	if acmePath == "" {
		acmePath = "."
	}

	// register all rules
	rule.Register(
		&rule.PathReplaceRegistry{},
		// next rule here
	)

	dockerWatcher := docker.New(docker.DefaultClient, docker.DefaultAddr)
	engine := engine.New(dockerWatcher)

	var server server

	if acmeEnable {
		server = acme.NewServer(engine, acmePath)
	} else {
		server = serverFun(func(handler http.Handler) error {
			return http.ListenAndServe(":80", engine)
		})
	}

	err := server.Start(engine)
	if err != nil {
		panic(err)
	}
}
