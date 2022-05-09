```
   ____        _
  | __ )  __ _| | _____ _ __      __ _  ___
  |  _ \ / _  | |/ / _ \ '__|    / _  |/ _ \
  | |_) | (_| |   <  __/ |   _  | (_| | (_) |
  |____/ \__,_|_|\_\___|_|  (_)  \__, |\___/
                                 |___/
```

# Introduction

Baker.go is a dynamic http reverse proxy designed to be highly extensible.

# Features

- [x] Include Docker driver to listen to Docker's events
- [x] Has exposed a driver interface which can be easily hook to other orchestration engines
- [x] Dynamic configuration, no need to restart reverse proxy in order to change the configuration
- [x] Uses a custom trie data structure, to compute fast path pattern matching
- [x] It can be use as library as it has implemented `http.Handler` interface
- [x] Highly extendable as most of the components have exposed interfaces
- [x] Middleware like feature to change the incoming and outgoing traffics
- [x] load balancing by default
- [x] Automatically updates and creates SSL certificates using `Let's Encrypt`

# Usage

First we need to run Baker inside docker. The following `docker-compose.yml`

```yml
version: "3.5"

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
      - "80:80"
      - "443:443"

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
version: "3.5"

services:
  service1:
    image: service:latest

    labels:
      - "baker.enable=true"
      - "baker.network=baker_net"
      - "baker.service.port=8000"
      - "baker.service.ping=/config"

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
    "rules": [
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

# Middleware

At the moment, there are 2 middlewares provided by default

### ReplacePath

Remove a specific path from incoming request. Service will be receiving the modified path.

in order to use this middleware, simply add the following rule to rules section of the configuration

```json
{
  "name": "ReplacePath",
  "config": {
    "search": "/sample1",
    "replace": "",
    "times": 1
  }
}
```

### AppendPath

Add a path at the beginning and end of path

in order to use this middleware, simply add the following rule to rules section of the configuration

```json
{
  "name": "AppendPath",
  "config": {
    "begin": "/begin",
    "end": "/end"
  }
}
```
