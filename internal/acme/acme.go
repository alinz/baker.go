package acme

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// PolicyManager is a
type PolicyManager interface {
	HostPolicy(ctx context.Context, host string) error
}

// Server contains all logic to handle both on port http and https
type Server struct {
	httpSrv       *http.Server
	httpsSrv      *http.Server
	policyManager PolicyManager
	certPath      string
}

// Start starts both http and https servers and initialize acme object
// NOTE: this methid is a blocking call, either run this as last statement or
// run it with a go command
func (s *Server) Start(handler http.Handler) error {

	errChan := make(chan error, 2)
	httpCloseSignal := make(chan struct{})
	httpsCloseSignal := make(chan struct{})

	manager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: s.policyManager.HostPolicy,
		Cache:      autocert.DirCache(s.certPath),
	}

	go func() {
		// start http server

		s.httpSrv.Handler = manager.HTTPHandler(nil)

		// logger.Info("acme http server is started")

		select {
		case errChan <- s.httpSrv.ListenAndServe():
			close(httpsCloseSignal)
		case <-httpsCloseSignal:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.httpSrv.Shutdown(ctx)
		}
	}()

	go func() {
		// start https server

		s.httpsSrv.TLSConfig = &tls.Config{GetCertificate: manager.GetCertificate}
		s.httpsSrv.Handler = handler

		// logger.Info("acme https server is started")

		select {
		case errChan <- s.httpsSrv.ListenAndServeTLS("", ""):
			close(httpsCloseSignal)
		case <-httpCloseSignal:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.httpsSrv.Shutdown(ctx)
		}
	}()

	// wait for one go routine to exit
	select {
	case <-httpCloseSignal:
	case <-httpsCloseSignal:
	}

	// this makes sure that we will return a value back
	// and not blocking the call
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// NewServer creates acme.Server
func NewServer(policyManager PolicyManager, certPath string) *Server {
	return &Server{
		httpSrv: &http.Server{
			Addr: ":80",
		},
		httpsSrv: &http.Server{
			Addr: ":443",
		},
		policyManager: policyManager,
		certPath:      certPath,
	}
}
