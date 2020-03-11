package main

import (
	"fmt"
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

var GitCommit string = "development"
var Version string = "master"

func main() {
	acmePath := os.Getenv("BAKER_ACME_PATH")
	acmeEnable := os.Getenv("BAKER_ACME") == "YES"

	if acmePath == "" {
		acmePath = "."
	}

	fmt.Fprintf(os.Stdout, `
  ____        _                                
 | __ )  __ _| | _____ _ __      __ _  ___   
 |  _ \ / _  | |/ / _ \ '__|    / _  |/ _ \  
 | |_) | (_| |   <  __/ |   _  | (_| | (_) | 
 |____/ \__,_|_|\_\___|_|  (_)  \__, |\___/  
                                |___/
Version: %s
Git Hash: %s 
https://github.com/alinz/baker.go

`, Version, GitCommit)

	// register all rules
	rule.Register(
		&rule.PathReplaceRegistry{},
		// next rule here
	)

	// initialize custim watcher
	dockerWatcher := docker.New(docker.DefaultClient, docker.DefaultAddr)
	err := dockerWatcher.Start()
	if err != nil {
		panic(err)
	}

	// initialize engine based on watcher
	engine := engine.New(dockerWatcher)

	// start the engine, if anything goes wrong, this method will
	// panic
	engine.Start()

	var server server

	if acmeEnable {
		server = acme.NewServer(engine, acmePath)
	} else {
		server = serverFun(func(handler http.Handler) error {
			return http.ListenAndServe(":80", engine)
		})
	}

	err = server.Start(engine)
	if err != nil {
		panic(err)
	}
}
