package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/http/pprof"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/docker"
	"github.com/alinz/baker.go/middleware"
)

var Version = "master"
var GitCommit = "development"

func main() {
	fmt.Fprintf(os.Stdout, `
_____       __    
| __ )  __ _| | _____ _ __      __ _  ___
|  _ \ / _  | |/ / _ \ '__|    / _  |/ _ \
| |_) | (_| |   <  __/ |   _  | (_| | (_) |
|____/ \__,_|_|\_\___|_|  (_)  \__, |\___/ 
                               |___/
Version: %s
Git Hash: %s 
https://github.com/alinz/baker.go

`, Version, GitCommit)

	acmePath := os.Getenv("BAKER_ACME_PATH")
	acmeEnable := strings.ToLower(os.Getenv("BAKER_ACME")) == "yes"
	logLevel := strings.ToLower(os.Getenv("BAKER_LOG_LEVEL"))

	errs := make(chan error, 10)

	registry := baker.NewRegistry()

	watcher := docker.NewWatcher(docker.UnixClient())
	containers := watcher.Watch(errs)

	go ping(5*time.Second, containers, registry, errs)

	go logError(errs)

	go func() {
		for {
			<-time.After(10 * time.Second)
			runtime.GC()
		}
	}()

	r := http.NewServeMux()

	if logLevel == "debug" {
		r.HandleFunc("/debug/pprof/", pprof.Index)
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	r.Handle("/", router(registry))

	if acmeEnable {
		err := baker.StartAcme(r, acmePath)
		if err != nil {
			panic(err)
		}
	} else {
		err := http.ListenAndServe(":80", r)
		if err != nil {
			panic(err)
		}
	}
}

func logError(errs <-chan error) {
	for err := range errs {
		fmt.Println(err)
	}
}

func ping(timeout time.Duration, containers <-chan baker.Container, registor baker.ContainerRegistor, errs chan<- error) {
	containersMap := make(map[string]baker.Container)

	for {
		select {
		case container := <-containers:
			if container.Addr() == "" {
				delete(containersMap, container.ID())
				continue
			}
			containersMap[container.ID()] = container

		case <-time.After(timeout):
			for _, container := range containersMap {
				fetch, ok := container.(baker.EndpointsFetcher)
				if !ok {
					errs <- fmt.Errorf("container %s is not EndpointFetcher", container.ID())
					break
				}

				endpoints, err := fetch.FetchEndpoints()
				if err != nil {
					errs <- err
					break
				}

				for _, endpoint := range endpoints {
					registor.UpdateContainer(container, endpoint)
				}
			}
		}
	}
}

func router(registor baker.ContainerRegistor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := r.Host
		path := r.URL.Path

		container, endpoint := registor.FindContainer(domain, path)
		if container == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "service is not available"}`))
			return
		}

		proxy := &httputil.ReverseProxy{
			Director: func(r *http.Request) {
				r.URL.Scheme = "http"
				r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
				r.URL.Host = container.Addr()

				if _, ok := r.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					r.Header.Set("User-Agent", "")
				}
			},
		}

		middlewares, err := middleware.Parse(endpoint.Rules)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			})
			return
		}

		middleware.Apply(proxy, middlewares...).ServeHTTP(w, r)
	})
}
