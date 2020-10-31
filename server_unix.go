// +build darwin linux freebsd openbsd netbsd

package httpserver

import (
	"crypto/tls"
	"fmt"
	"net/http"

	_grace "github.com/facebookgo/grace/gracehttp"
)

func (s *Server) serve() error {
	var tlsConfig *tls.Config
	tlsConfig = s.tls
	return _grace.Serve(&http.Server{
		Addr:        fmt.Sprintf(":%d", s.port),
		Handler:     s.cors.Handler(s.handlers),
		IdleTimeout: s.idleTimeout,
		TLSConfig:   tlsConfig,
	})
}
