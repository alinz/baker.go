package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/driver/docker"
	"github.com/alinz/baker.go/pkg/acme"
	"github.com/alinz/baker.go/pkg/log"
	"github.com/alinz/baker.go/rule"
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

	log.Configure(log.Config{
		ConsoleLoggingEnabled: true,
		FileLoggingEnabled:    true,
		Directory:             "./logs",
		Filename:              "baker.log",
		Level:                 logLevel,
	})

	done := make(chan struct{}, 1)
	defer close(done)

	go func() {
		bToMb := func(b uint64) uint64 {
			return b / 1024 / 1024
		}

		if logLevel != "debug" {
			return
		}

		for {
			select {
			case <-done:
				return
			case <-time.After(10 * time.Second):
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				log.
					Debug().
					Uint64("Alloc", bToMb(m.Alloc)).
					Uint64("TotalAlloc", bToMb(m.TotalAlloc)).
					Uint64("Sys", bToMb(m.Sys)).
					Uint32("NumGC", m.NumGC).
					Msg("memory usage")
			}
		}
	}()

	containers, err := docker.New()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create docker driver")
	}

	baker := baker.New(
		containers,
		rule.RegisterAppendPath(),
		rule.RegisterReplacePath(),
	)

	if acmeEnable {
		err := acme.Start(baker, acmePath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to start acme")
		}
	} else {
		err := http.ListenAndServe(":80", baker)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}
}
