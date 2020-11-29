package httpserver

import (
	"fmt"
	"net/http"
	"time"
)

func (s *Server) log(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		elapsed := time.Since(start)
		var statusCode int
		rw, ok := w.(*responseWriter)
		if !ok { // impossible...!!! but let be safe.
			statusCode = http.StatusOK // default http.ResponseWriter status code
		} else {
			statusCode = rw.statusCode
		}
		if s.enableLogger {
			if statusCode >= 400 {
				s.logger.Printf("%s | httpserver | %s | %d | %s | %v | %s\n", time.Now().Format(time.RFC3339), r.Method, statusCode, r.URL.Path, elapsed, r.Header.Get("Request-Id"))
			} else {
				fmt.Printf("%s | httpserver | %s | %d | %s | %v | %s\n", time.Now().Format(time.RFC3339), r.Method, statusCode, r.URL.Path, elapsed, r.Header.Get("Request-Id"))
			}
		}
	}
}
