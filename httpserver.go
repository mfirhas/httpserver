package httpserver

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	_uuid "github.com/google/uuid"
	_router "github.com/julienschmidt/httprouter"
	_cors "github.com/rs/cors"
)

type Server struct {
	handlers    *_router.Router
	errChan     chan error
	port        uint16
	idleTimeout time.Duration
	logger      *log.Logger
	tls         *tls.Config
	cors        *_cors.Cors
	middlewares []Middleware

	panicHandler    PanicHandler
	notFoundHandler http.Handler
}

type Middleware func(next http.HandlerFunc, params ...interface{}) http.HandlerFunc
type PanicHandler func(w http.ResponseWriter, r *http.Request, rcv ...interface{})

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

	// PanicHandler triggered if panic happened.
	// rcv: first param is argument retrieved from `recover()` function.
	PanicHandler PanicHandler

	// NotFoundHandler triggered if path not found.
	// If empty then default is used.
	NotFoundHandler http.HandlerFunc
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
	var notFoundHandler http.Handler
	if opts.NotFoundHandler != nil {
		notFoundHandler = &notFound{opts.NotFoundHandler}
	}
	srv := &Server{
		handlers:        h,
		port:            opts.Port,
		idleTimeout:     opts.IdleTimeout,
		logger:          log.New(os.Stderr, "", 0),
		middlewares:     make([]Middleware, 0),
		tls:             opts.TLS,
		cors:            cors,
		errChan:         make(chan error),
		panicHandler:    opts.PanicHandler,
		notFoundHandler: notFoundHandler,
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

type notFound struct {
	handler http.HandlerFunc
}

func (n *notFound) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.handler(w, r)
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
	requestID  string
	xRequestID string
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func newResponseWriter(w http.ResponseWriter, reqID string, xReqID string) *responseWriter {
	// default if not set is 200
	return &responseWriter{w, http.StatusOK, reqID, xReqID}
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
		rw := newResponseWriter(w, r.Header.Get("Request-Id"), r.Header.Get("X-Request-Id"))
		next(rw, r)
	}
}

func (s *Server) recoverPanic(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rcv := recover(); rcv != nil {
				ResponseString(w, http.StatusInternalServerError, "httpserver got panic")
				s.logger.Printf("%s | httpserver | %s | %s | %s | %s\n", time.Now().Format(time.RFC3339), "PANIC", r.Method, r.URL.Path, r.Header.Get("Request-Id"))
				s.logger.Printf("☠️ ☠️ ☠️ ☠️ ☠️ ☠️  PANIC START (%s) ☠️ ☠️ ☠️ ☠️ ☠️ ☠️", r.Header.Get("Request-Id"))
				debug.PrintStack()
				s.logger.Printf("☠️ ☠️ ☠️ ☠️ ☠️ ☠️  PANIC END (%s) ☠️ ☠️ ☠️ ☠️ ☠️ ☠️", r.Header.Get("Request-Id"))
				if s.panicHandler != nil {
					s.panicHandler(w, r, rcv)
				}
				return
			}
		}()
		next(w, r)
	}
}

func (s *Server) GET(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.GET(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
}

func (s *Server) HEAD(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.HEAD(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
}

func (s *Server) POST(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.POST(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
}

func (s *Server) PUT(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	s.handlers.PUT(path, f(s.recoverPanic(s.chainMiddlewares(handler, middlewares...))))
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
