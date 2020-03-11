package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/alinz/baker.go/driver/docker"
	"github.com/alinz/baker.go/engine"
	"github.com/alinz/baker.go/internal/acme"
	"github.com/alinz/baker.go/pkg/logger"
	"github.com/alinz/baker.go/pkg/rule"
)

type server interface {
	Start(handler http.Handler) error
}

type serverFun func(handler http.Handler) error

func (s serverFun) Start(handler http.Handler) error {
	return s(handler)
}

// GitCommit will be set by build scriot
var GitCommit string = "development"

// Version will be set by build script and refer to tag version
var Version string = "master"

func setLogLevel(val string) {
	var level logger.Level

	switch val {
	case "all":
		level = logger.All
	case "debug":
		level = logger.Debug
	case "info":
		level = logger.Info
	case "warn":
		level = logger.Warn
	case "error":
		level = logger.Error
	default:
		level = logger.Info
	}

	logger.Default.Level(level)
}

func main() {
	acmePath := os.Getenv("BAKER_ACME_PATH")
	acmeEnable := strings.ToLower(os.Getenv("BAKER_ACME")) == "yes"
	logLevel := strings.ToLower(os.Getenv("BAKER_LOG_LEVEL"))

	setLogLevel(logLevel)

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
