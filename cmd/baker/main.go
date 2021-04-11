package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/internal/logger"
)

// GitCommit will be set by build scriot
var GitCommit string = "development"

// Version will be set by build script and refer to tag version
var Version string = "master"

func main() {
	acmePath := os.Getenv("BAKER_ACME_PATH")
	acmeEnable := strings.ToLower(os.Getenv("BAKER_ACME")) == "yes"
	logLevel := strings.ToLower(os.Getenv("BAKER_LOG_LEVEL"))

	switch strings.ToUpper(logLevel) {
	case "ALL":
		logger.Default.Level(logger.AllLevel)
	case "DEBUG":
		logger.Default.Level(logger.DebugLevel)
	case "ERROR":
		logger.Default.Level(logger.ErrorLevel)
	case "WARN":
		logger.Default.Level(logger.WarnLevel)
	case "INFO":
		fallthrough
	default:
		logger.Default.Level(logger.InfoLevel)
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

	watcher := baker.NewDockerWatcher(baker.DefaultDockerWatcherConfig)
	pinger := baker.NewBasePinger(watcher)
	store := baker.NewBaseStore(pinger)
	router := baker.NewBaseRouter(store)

	router.
		AddProcessor("ReplacePath", baker.CreateProcessorPathReplace).
		AddProcessor("AppendPath", baker.CreateProcessorPathAppend)

	r := http.NewServeMux()
	r.HandleFunc("/", router.ServeHTTP)

	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	simpleRouter := &SimpleRouter{
		r: r,
	}

	go watcher.Start()

	if acmeEnable {
		if acmePath == "" {
			acmePath = "."
		}

		acme := baker.NewAcmeServer(simpleRouter, acmePath)
		acme.Start(router)
	} else {
		http.ListenAndServe(":80", simpleRouter.r)
	}
}

type SimpleRouter struct {
	r http.Handler
}

func (s *SimpleRouter) HostPolicy(ctx context.Context, host string) error {
	return nil
}
