package httpserver

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	_uuid "github.com/google/uuid"
	_router "github.com/julienschmidt/httprouter"
	_cors "github.com/rs/cors"
)

// TODO add SO_REUSEPORT support

type Server struct {
	handlers    *_router.Router
	errChan     chan error
	port        uint16
	idleTimeout time.Duration
	logger      *log.Logger
	tls         *tls.Config
	cors        *_cors.Cors
	middlewares []Middleware
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

type Opts struct {
	Port uint16

	// EnableLogger enable logging for incoming requests
	EnableLogger bool

	// IdleTimeout keep-alive timeout while waiting for the next request coming. If empty then no timeout.
	IdleTimeout time.Duration

	// TLS to enable HTTPS
	TLS *tls.Config

	// Cors optional, can be nil, if nil then default will be set.
	Cors *Cors
}

// Cors corst options
type Cors struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	MaxAge           int
	AllowCredentials bool
	IsDebug          bool
}

func New(opts *Opts) *Server {
	h := _router.New()
	var cors *_cors.Cors
	if opts.Cors != nil {
		cors = _cors.New(_cors.Options{
			AllowedOrigins:     opts.Cors.AllowedOrigins,
			AllowedMethods:     opts.Cors.AllowedMethods,
			AllowedHeaders:     opts.Cors.AllowedHeaders,
			ExposedHeaders:     opts.Cors.ExposedHeaders,
			MaxAge:             opts.Cors.MaxAge,
			AllowCredentials:   opts.Cors.AllowCredentials,
			OptionsPassthrough: true,
			Debug:              opts.Cors.IsDebug,
		})
	}
	srv := &Server{
		handlers:    h,
		port:        opts.Port,
		idleTimeout: opts.IdleTimeout,
		logger:      log.New(os.Stderr, "", 0),
		middlewares: make([]Middleware, 0),
		tls:         opts.TLS,
		cors:        cors,
		errChan:     make(chan error),
	}
	if opts.EnableLogger {
		w := make(buffer, 10<<20)
		go write(w)
		srv.logger = log.New(w, "", 0)
		srv.middlewares = append(srv.middlewares, srv.log)
	}
	return srv
}

// Run the server. Blocking.
func (s *Server) Run() {
	s.logger.Printf("%s | httpserver | server is starting...", time.Now().Format(time.RFC3339))
	s.logger.Printf("%s | httpserver | server is running on port %d", time.Now().Format(time.RFC3339), s.port)
	if err := s.serve(); err != nil {
		s.logger.Printf("%s | httpserver | server failed with error: %v", time.Now().Format(time.RFC3339), err)
		s.errChan <- err
	}
}

func (s *Server) ListenError() <-chan error {
	return s.errChan
}

// TLSConfig generate certificate config using provided certificate and private key.
// It will overwrite the one set in Opts.
func (s *Server) TLSConfig(cert, key string) error {
	certificate, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return err
	}
	s.tls = &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}
	return nil
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	// default if not set is 200
	return &responseWriter{w, http.StatusOK}
}

func f(next http.HandlerFunc) _router.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps _router.Params) {
		if r.Header.Get("Request-Id") == "" && r.Header.Get("X-Request-Id") == "" {
			r.Header.Set("Request-Id", _uuid.New().String())
		}
		if r.Header.Get("Request-Id") == "" && r.Header.Get("X-Request-Id") != "" {
			r.Header.Set("Request-Id", r.Header.Get("X-Request-Id"))
		}
		if len(ps) > 0 {
			urlValues := r.URL.Query()
			for i := range ps {
				urlValues.Add(ps[i].Key, ps[i].Value)
			}
			r.URL.RawQuery = urlValues.Encode()
		}
		rw := newResponseWriter(w)
		next(rw, r)
	}
}

func (s *Server) recoverPanic(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ResponseString(w, r, http.StatusInternalServerError, "httpserver got panic")
				s.logger.Printf("%s | httpserver | %s | %s | %s | %s\n", time.Now().Format(time.RFC3339), "PANIC", r.Method, r.URL.Path, r.Header.Get("Request-Id"))
				s.logger.Printf("☠️ ☠️ ☠️ ☠️ ☠️ ☠️  PANIC START (%s) ☠️ ☠️ ☠️ ☠️ ☠️ ☠️", r.Header.Get("Request-Id"))
				debug.PrintStack()
				s.logger.Printf("☠️ ☠️ ☠️ ☠️ ☠️ ☠️  PANIC END (%s) ☠️ ☠️ ☠️ ☠️ ☠️ ☠️", r.Header.Get("Request-Id"))
				return
			}
		}()
		next(w, r)
	}
}

func responseHeader(w http.ResponseWriter, r *http.Request, statusCode int) {
	w.Header().Set("Date", time.Now().Format(time.RFC1123))
	w.Header().Set("Request-Id", r.Header.Get("Request-Id"))
	w.Header().Set(fmt.Sprintf("X-Req-Id_%s-Status_code", r.Header.Get("Request-Id")), strconv.Itoa(statusCode))
	w.WriteHeader(statusCode)
}

// Response response by writing the body to http.ResponseWriter.
// Call at the end line of your handler.
func Response(w http.ResponseWriter, r *http.Request, statusCode int, body []byte) {
	responseHeader(w, r, statusCode)
	w.Write(body)
}

// ResponseJSON response by writing body with json encoder into http.ResponseWriter.
// Body must be either struct or map[string]interface{}. Otherwise would result in incorrect parsing at client side.
// If you have []byte as response body, then use Response function instead.
// Call at the end line of your handler.
func ResponseJSON(w http.ResponseWriter, r *http.Request, statusCode int, body interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	responseHeader(w, r, statusCode)
	return json.NewEncoder(w).Encode(body)
}

// ResponseString response in form of string whatever passed into body param.
// Call at the end line of your handler.
func ResponseString(w http.ResponseWriter, r *http.Request, statusCode int, body interface{}) {
	responseHeader(w, r, statusCode)
	fmt.Fprintf(w, "%v", body)
}

func (s *Server) GET(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.GET(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
}

func (s *Server) HEAD(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.HEAD(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
}

func (s *Server) HEADGET(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.HEAD(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
	s.handlers.GET(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
}

func (s *Server) POST(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.POST(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
}

func (s *Server) PUT(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.POST(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
}

func (s *Server) DELETE(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.DELETE(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
}

func (s *Server) PATCH(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.PATCH(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
}

func (s *Server) OPTIONS(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.OPTIONS(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
}

// FILES serve files from 1 directory dynamically.
// @filePath: must end with '/*filepath' as placeholder for filename to be accessed.
// @rootPath: root directory where @filepath locate.
func (s *Server) FILES(filePath string, rootPath string, middlewares ...Middleware) {

	if len(filePath) < 10 || filePath[len(filePath)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + filePath + "'")
	}

	rootDir := http.Dir(rootPath)
	fileServer := http.FileServer(rootDir)

	s.GET(filePath, func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = r.URL.Query().Get("filepath")
		fileServer.ServeHTTP(w, r)
	}, middlewares...)
}
