// +build windows

package httpserver

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

// graceful is not support in Windows. Using built-in package instead. This is for avoiding this package failed to run locally, rarely Windows used in server now.
func (s *Server) serve() error {
	var handler http.Handler = s.handlers
	if s.cors != nil {
		handler = s.cors.Handler(s.handlers)
	}
	var tlsConfig *tls.Config
	tlsConfig = s.tls
	if s.notFoundHandler != nil {
		s.handlers.NotFound = s.notFoundHandler
	}
	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", s.port),
		Handler:     handler,
		IdleTimeout: s.idleTimeout,
		TLSConfig:   tlsConfig,
	}

	if tlsConfig != nil {
		return srv.ListenAndServeTLS("", "")
	}
	return srv.ListenAndServe()
}
