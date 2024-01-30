```
   ____        _
  | __ )  __ _| | _____ _ __      __ _  ___
  |  _ \ / _  | |/ / _ \ '__|    / _  |/ _ \
  | |_) | (_| |   <  __/ |   _  | (_| | (_) |
  |____/ \__,_|_|\_\___|_|  (_)  \__, |\___/
                                 |___/
```

# Introduction

Baker.go is a dynamic HTTP reverse proxy with a focus on extensibility and flexibility. It is designed to adapt to a variety of orchestration engines and provides dynamic configuration capabilities, eliminating the need to restart the reverse proxy when changing configurations.

# Features

- Docker driver integration for Docker event listening.
- Exposed driver interface for easy integration with other orchestration engines.
- Dynamic configuration capabilities.
- Custom trie data structure for fast path pattern matching.
- Can be used as a library, as it implements the HTTP Handler interface.
- High extensibility due to exposed interfaces for most components.
- Middleware-like feature for modifying incoming and outgoing traffic.
- Default load balancing.
- Automatic SSL certificate updates and creation using Let's Encrypt.
- Configurable rate limiter per domain and path.

# Usage

First, we need to run Baker inside docker. The following `docker-compose.yml`

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

Then for each service, the following `docker-compose` can be used. The only requirements are labels and networks. Make sure both baker and service have the same network interface

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

The service should expose a REST endpoint that returns a configuration. This endpoint acts as a health check and provides real-time configuration.

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
        "type": "ReplacePath",
        "args": {
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

Baker.go comes with several built-in middleware:

### ReplacePath

Remove a specific path from an incoming request. Service will be receiving the modified path.
to use this middleware, simply add the following rule to the rules section of the configuration

```json
{
  "type": "ReplacePath",
  "args": {
    "search": "/sample1",
    "replace": "",
    "times": 1
  }
}
```

### AppendPath

Add a path at the beginning and end of the path
to use this middleware, simply add the following rule to the rules section of the configuration

```json
{
  "type": "AppendPath",
  "args": {
    "begin": "/begin",
    "end": "/end"
  }
}
```

### RateLimiter

Add a rate limiter for a specific domain and path
to use this middleware, simply add the following rule to the riles sections of the configuration

```json
{
  "type": "RateLimiter",
  "args": {
    "request_limit": 100,
    "window_duration": "60s"
  }
}
```

the above configuration means, in one minute, 100 requests should be routed per individual IP address, if that is exceeded, a 429 HTTP status will be sent back to the client.

## License

Baker.go is licensed under the [MIT License](LICENSE.md).
