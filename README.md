```
   ____        _                                
  | __ )  __ _| | _____ _ __      __ _  ___   
  |  _ \ / _  | |/ / _ \ '__|    / _  |/ _ \  
  | |_) | (_| |   <  __/ |   _  | (_| | (_) | 
  |____/ \__,_|_|\_\___|_|  (_)  \__, |\___/  
                                 |___/
```

# Introduction

Baker.go is a dynamic http reverse proxy with battery included. It has docker driver to listen to events coming from docker and register services to receive incoming traffic.

# Features

Baker.go was rewritten based on previous [baker](https://github.com/alinz/baker). The code is simpler and more easy to be modified.

Heare are couple of main features of `Baker.go`:

- [x] Docker driver to listen to Docker's events
- [x] Has exposed a driver interface which can be easily hook to other orchestration engines
- [x] Dyanmic configurations, no need to restrat reverse proxy in order to change the configration
- [x] Uses only Standard Library
- [x] It can be use as library as it has implemented `http.Handler` interface
- [x] Highly extendable as most of the componets have exposed interfaces 
- [x] Middleware like feature to change the incoming traffic
- [x] Round-Robin loadbalancing by default
- [x] Automatically updates and creates SSL certificates using `Let's Encrypt`

# Usage

In order to use `Baker.go`, it needs to be accessed to `/var/run/docker.sock` file descriptor. The following `docker-compose` file contains enough logic to keep it started:

```yml
version: '3.5'

services:
  service1:
    image: alinz/baker.go:latest

    environment:
      # enables ACME system
      - BAKER_ACME=NO
      # folder location which holds all certification
      - BAKER_ACME_PATH=/acme/cert
      # log level of baker. Debug, Warn, Info, Error are the values for level
      - BAKER_LOG_LEVEL=Debug

    ports:
      - '80:80'
      - '443:443' # if BAKER_ACME set to YES, this port needs to be available

    networks:
      - baker

    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./acme/cert:/acme/cert

networks:
  baker:
    name: baker_net
    driver: bridge
```

if automatice SSL certificate is desiered, `BAKER_ACME` should be set to `YES`. Once the `docker-compose.yml` is launches, the following content message should be displayed:

```bash
service1_1  | 
service1_1  |   ____        _                                
service1_1  |  | __ )  __ _| | _____ _ __      __ _  ___   
service1_1  |  |  _ \ / _  | |/ / _ \ '__|    / _  |/ _ \  
service1_1  |  | |_) | (_| |   <  __/ |   _  | (_| | (_) | 
service1_1  |  |____/ \__,_|_|\_\___|_|  (_)  \__, |\___/  
service1_1  |                                 |___/
service1_1  | Version: vx.x.x
service1_1  | Git Hash: xxxxxxxxxxxxxxxxxxxxx
service1_1  | https://github.com/alinz/baker.go
```

> NOTE: depends on log level, the content of stdout should be different.

Since `Baker.go` listens to docker's events, any containers that deploy to docker engine will be registered to `Baker.go`.
In order to configure the registration process, docker's label is being used.

There are 3 custom labels which described all requirements:

- baker.enable

  > containers need to explicitly say that they want to be registered at `Baker.go`

- baker.network

  > describes which network name `Baker.go` is running. This is essential for baker to route the traffic

- baker.service.port

  > describes which ports this container is listening to

- baker.service.ping

  > describes which path `Baker.go` should use to check the health check and get the dynamic configuration

A sample example of `docker-compose.yml` for a service is as follows:

```yml
version: '3.5'

services:
  service1:
    image: baker/service:latest

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

Now, the service needs to have a custom `/config` path that returns the following JSON, this path will be hit by `Baker.go` every 10 seconds. 

```json
[
  {
    "domain": "api.example.com",
    "path": "/v1*",
    "ready": true,
    "rules": [
      {
        "name": "path_replace",
        "config": {
          "search": "/v1",
          "replace": "",
          "times": 1
        }
      }
    ]
  },
  {
    "domain": "example.com",
    "path": "/*",
    "ready": true
  }  
]
```

This configuration tells `Baker.go` that there are 2 domains, `api.example.com/v1*` and `example.com/*` which I want to receive the incoming traffic from.
The rules also saying that if a request comming as `api.example.com/v1/users/1`, replace `/v1` to `/`. So the container will receive the request path as `/users/1`

# As Library

Since `Baker.go` has been implemented `http.Hanlder` interface, it can be used inside your go application as library.

Here are the things need to be done:

First you need to registered all rules that your application needed. 

```go
  import "github.com/alinz/baker.go/pkg/rule"

	// register all rules
	rule.Register(
		&rule.PathReplaceRegistry{},
		// next rule here
	)
```

A custom rule can be created by creating 2 structs, one for registration and one for execution.

```go

// Registrar interface for registring your rule with `rule.Register`
type Registrar interface {
	Name() string
	CreateHandler() Handler
}

// Hanlder is the execution of rule. Everytime a rule loaded via config path, it first validates by calling Valid method
type Handler interface {
	Valid() error
	ApplyRule(r *http.Request)
}

```

Once rules have been registred, a watcher needs to be initialized. Watcher is an interface from `driver.Watcher`

```go

// Watcher defines how driver should react
type Watcher interface {
	// Container gets the next availabel container
	// this is a blocking calls
	Container() *baker.Container
}

```

Here is an example of docker's watcher:

```go

// initialize custim watcher
dockerWatcher := docker.New(docker.DefaultClient, docker.DefaultAddr)
err := dockerWatcher.Start()
if err != nil {
	panic(err)
}
  
```

once a watcher is started, an engine can be created


```go

// initialize engine based on watcher
engine := engine.New(dockerWatcher)
// start the engine, if anything goes wrong, this method will
// panic
engine.Start()
  
```

Engine is implemented `http.Handler` interface and can be used as a handler to a custom routing library.
