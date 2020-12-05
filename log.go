package httpserver

import (
	"bufio"
	"net/http"
	"os"
	"time"
)

// buffer memory to store log before writing them into writer/file
type buffer chan []byte

// Write overwrite io.Writer Write method to instead of writing directly into file,
// it passes the bytes into buffer memory to be written into actual writer/file asynchronously.
func (b buffer) Write(p []byte) (int, error) {
	b <- append(([]byte)(nil), p...)
	return len(p), nil
}

// worker to write log data from buffer memory into writer/file asynchronously.
func write(b buffer) {
	writer := bufio.NewWriter(os.Stderr)
	for p := range b {
		writer.Write(p)
		writer.Flush()
	}
}

// middleware for log
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
		s.logger.Printf("%s | httpserver | %s | %d | %s | %v | %s\n", time.Now().Format(time.RFC3339), r.Method, statusCode, r.URL.Path, elapsed, r.Header.Get("Request-Id"))
	}
}
