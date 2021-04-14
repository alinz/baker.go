package baker

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

func StartAcme(handler http.Handler, cachePath string) error {
	if cachePath == "" {
		cachePath = "."
	}

	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(cachePath),
	}

	httpsServer := &http.Server{
		Addr:         ":443",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	httpServer := &http.Server{
		Addr:         ":80",
		Handler:      certManager.HTTPHandler(nil),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	httpClose := make(chan struct{}, 1)
	httpsClose := make(chan struct{}, 1)
	errs := make(chan error, 2)

	go func() {
		defer close(httpClose)
		errs <- httpServer.ListenAndServe()
	}()

	go func() {
		defer close(httpsClose)
		errs <- httpsServer.ListenAndServeTLS("", "")
	}()

	select {
	case <-httpClose:
		httpsServer.Shutdown(context.Background())
	case <-httpsClose:
		httpServer.Shutdown(context.Background())
	}

	select {
	case err := <-errs:
		return err
	default:
		return nil
	}
}
