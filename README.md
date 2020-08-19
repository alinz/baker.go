```
   ____        _                                
  | __ )  __ _| | _____ _ __      __ _  ___   
  |  _ \ / _  | |/ / _ \ '__|    / _  |/ _ \  
  | |_) | (_| |   <  __/ |   _  | (_| | (_) | 
  |____/ \__,_|_|\_\___|_|  (_)  \__, |\___/  
                                 |___/
```

# Introduction

Baker.go is a dynamic http reverse proxy designed to be highly extensible. It has 4 major parts.

<img src="./doc/baker.svg">

- Watcher

listen to orchestration system and produce container object

```go
// Watcher defines how driver should react
type Watcher interface {
	// Container gets the next available container
	// this is a blocking calls, if Container object is nil
	// it means that Watcher has been closed
	Container() (*Container, error)
}
```

- Pinger

consumes container objects from Watcher and ping each container for health check and retrieve configuration and produce Service object

```go
// Pinger defines how container can be ping to get extra information
// and construct Service object
type Pinger interface {
	// Service if service returns an nil value, it means that Pinger has
	// some internal error and it will no longer return service
	Service() (*Service, error)
}
```

- Store

consumes service objects from Pinger and store them in highly optimize trie data structure for fast access

```go
// Store will be used to get Container
type Store interface {
	// Query returns a service based on domain and path
	Query(domain, path string) *Service
}
```

- Router

Router uses Store to query each service, extract recipes and proxy requests to each container. Router is http.Handler compatible.



# Features

- [x] Include Docker driver to listen to Docker's events
- [x] Has exposed a driver interface which can be easily hook to other orchestration engines
- [x] Dynamic configuration, no need to restart reverse proxy in order to change the configuration
- [x] Uses only Standard Library
- [x] Uses a custom Trie data structure, to compute fast path pattern matching
- [x] It can be use as library as it has implemented `http.Handler` interface
- [x] Highly extendable as most of the components have exposed interfaces 
- [x] Middleware like feature to change the incoming traffic, outgoing traffic and custom error handling
- [x] Round-Robin load balancing by default
- [x] Automatically updates and creates SSL certificates using `Let's Encrypt`


# Usage

First we need to run Baker inside docker. The following `docker-compose.yml`


```yml
version: '3.5'

services:
  baker:
    image: alinz/baker.go:latest

    environment:
      # enables ACME system
      - BAKER_ACME=NO
      # folder location which holds all certification
      - BAKER_ACME_PATH=/acme/cert
      - BAKER_LOG_LEVEL=DEBUG

    ports:
      - '80:80'
      - '443:443'

    # make sure to use the right network
    networks:
      - baker

    volumes:
      # make sure it can access to main docker.sock
      - /var/run/docker.sock:/var/run/docker.sock
      - ./acme/cert:/acme/cert

networks:
  baker:
    name: baker_net
    driver: bridge
```


Then for each service, a following `docker-compose` can be used. The only requirements is labels and networks. Make sure both baker and service has the same network interface

```yml
version: '3.5'

services:
  service1:
    image: service:latest

    labels:
      - 'baker.enable=true'
      - 'baker.network=baker_net'
      - 'baker.service.port=8000'
      - 'baker.service.ping=/config'

    networks:
      - baker

networks:
  baker:
    external:
      name: baker_net    
```

The service, should expose a REST endpoint which returns a configuration, the configuration endpoint act as a health check and providing realtime configuration:

```json
[
  {
    "domain": "example.com",
    "path": "/sample1",
    "ready": true
  },
  {
    "domain": "example.com",
    "path": "/sample2",
    "ready": false
  },
  {
    "domain": "example1.com",
    "path": "/sample1*",
    "ready": true,
    "recipes": [
      {
        "name": "ReplacePath",
        "config": {
          "search": "/sample1",
          "replace": "",
          "times": 1
        }
      }
    ]
  }
]
```

# Custom Baker

Baker is very extensible and easy to build. If a new recipe needed to be added to Router, it can be done in few lines of code as describe in following code, or can be found in `cmd/baker/main.go`

```go
package main

import (
	"net/http"

	"github.com/alinz/baker.go"
)

func main() {
  watcher := baker.NewDockerWatcher(baker.DefaultDockerWatcherConfig)
  pinger := baker.NewBasePinger(watcher)
  store := baker.NewBaseStore(pinger)
  router := baker.NewBaseRouter(store)

  // add all recipes and processors with a unique names here
  router.AddProcessor("ReplacePath", baker.CreateProcessorPathReplace)

  go watcher.Start()

  http.ListenAndServe(":80", router)
}
```