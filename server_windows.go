// +build windows

package httpserver

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

// graceful is not support in Windows. Using built-in package instead. This is for avoiding this package failed to run locally, rarely Windows used in server now.
func (s *Server) serve() error {
	var tlsConfig *tls.Config
	tlsConfig = s.tls
	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", s.port),
		Handler:     s.cors.Handler(s.handlers),
		IdleTimeout: s.idleTimeout,
		TLSConfig:   tlsConfig,
	}

	if tlsConfig != nil {
		return srv.ListenAndServeTLS("", "")
	}
	return srv.ListenAndServe()
}
