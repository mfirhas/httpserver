// +build darwin linux freebsd openbsd netbsd

package httpserver

import (
	"crypto/tls"
	"fmt"
	"net/http"

	_grace "github.com/facebookgo/grace/gracehttp"
)

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
	return _grace.Serve(&http.Server{
		Addr:        fmt.Sprintf(":%d", s.port),
		Handler:     handler,
		IdleTimeout: s.idleTimeout,
		TLSConfig:   tlsConfig,
	})
}
