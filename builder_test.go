package httpserver

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	_router "github.com/julienschmidt/httprouter"
	_cors "github.com/rs/cors"
)

var (
	port = uint16(2000)
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestBuild(t *testing.T) {
	testSB := Build(port)
	expectedServer := &ServerBuilder{
		srv: &Server{
			port:     port,
			handlers: _router.New(),
			logger:   log.New(os.Stderr, "", 0),
		},
	}
	fmt.Println(expectedServer)
	fmt.Println(testSB)
	if !reflect.DeepEqual(expectedServer, testSB) {
		t.Errorf("error: expected %v, got %v", expectedServer, testSB)
	}
}

func TestWithIdleTimeout(t *testing.T) {
	testSB := Build(port)
	idleTimeout := time.Duration(2)
	expectedServer := &ServerBuilder{
		srv: &Server{
			port:        port,
			handlers:    _router.New(),
			idleTimeout: idleTimeout,
			logger:      log.New(os.Stderr, "", 0),
		},
	}
	sb := testSB.WithIdleTimeout(idleTimeout)
	if !reflect.DeepEqual(expectedServer, sb) {
		t.Errorf("error: expected %v, got %v", expectedServer, sb)
	}
}

func TestWithCors(t *testing.T) {
	testSB := Build(port)
	cors := &Cors{
		AllowCredentials: true,
		AllowedHeaders:   []string{"testHeader"},
		AllowedMethods:   []string{"POST", "GET"},
		MaxAge:           123003,
	}
	c := _cors.New(_cors.Options{
		AllowedOrigins:     cors.AllowedOrigins,
		AllowedMethods:     cors.AllowedMethods,
		AllowedHeaders:     cors.AllowedHeaders,
		ExposedHeaders:     cors.ExposedHeaders,
		MaxAge:             cors.MaxAge,
		AllowCredentials:   cors.AllowCredentials,
		OptionsPassthrough: true,
		Debug:              cors.IsDebug,
	})
	expectedServer := &ServerBuilder{
		srv: &Server{
			port:     port,
			handlers: _router.New(),
			cors:     c,
			logger:   log.New(os.Stderr, "", 0),
		},
	}
	sb := testSB.WithCors(cors)
	if !reflect.DeepEqual(expectedServer, sb) {
		t.Errorf("error: expected %v, got %v", expectedServer, sb)
	}
}

func TestWithLogger(t *testing.T) {
	testSB := Build(port)
	var middlewares []Middleware
	middlewares = append(middlewares, testSB.srv.log)
	expectedServer := &ServerBuilder{
		srv: &Server{
			port:        port,
			handlers:    _router.New(),
			middlewares: middlewares,
			logger:      log.New(os.Stderr, "", 0),
		},
	}

	sb := testSB.WithLogger()
	if sb.srv.logger == nil || (len(sb.srv.middlewares) != 1 && !reflect.DeepEqual(sb.srv.middlewares[0], expectedServer.srv.middlewares[0])) {
		t.Errorf("error: expected %v, got %v", sb.srv.middlewares[0], expectedServer.srv.middlewares[0])
	}
}

func TestWithTLS(t *testing.T) {
	testSB := Build(port)
	tls := &tls.Config{}
	expectedServer := &ServerBuilder{
		srv: &Server{
			port:     port,
			handlers: _router.New(),
			tls:      tls,
			logger:   log.New(os.Stderr, "", 0),
		},
	}
	sb := testSB.WithTLS(tls)
	if !reflect.DeepEqual(expectedServer, sb) {
		t.Errorf("error: expected %v, got %v", expectedServer, sb)
	}
}

func TestWithPanicHandler(t *testing.T) {
	testSB := Build(port)
	panicHandler := func(w http.ResponseWriter, r *http.Request, rcv ...interface{}) {}
	sb := testSB.WithPanicHandler(panicHandler)
	if sb.srv.panicHandler == nil {
		t.Errorf("error: expected not null")
	}
}

func TestWithNotFoundHandler(t *testing.T) {
	testSB := Build(port)
	notfoundHandler := func(w http.ResponseWriter, r *http.Request) {}
	sb := testSB.WithNotFoundHandler(notfoundHandler)
	if sb.srv.notFoundHandler == nil {
		t.Errorf("error: expected not null")
	}
}

func TestWithMiddleware(t *testing.T) {
	testSB := Build(port)
	m := func(next http.HandlerFunc, params ...interface{}) http.HandlerFunc { return nil }
	var middlewares []Middleware
	middlewares = append(middlewares, m)
	sb := testSB.WithMiddleware(m)
	if len(sb.srv.middlewares) != 1 {
		t.Errorf("error: expected contain 1 middleware")
	}
}

func TestAddHandler(t *testing.T) {
	testSB := Build(port)

	handlerGet := func(w http.ResponseWriter, r *http.Request) {}
	testSB.AddHandler(http.MethodGet, "/path", handlerGet)
	handleGet, _, _ := testSB.srv.handlers.Lookup(http.MethodGet, "/path")
	if handleGet == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerPost := func(w http.ResponseWriter, r *http.Request) {}
	testSB.AddHandler(http.MethodPost, "/path", handlerPost)
	handlePost, _, _ := testSB.srv.handlers.Lookup(http.MethodPost, "/path")
	if handlePost == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerHead := func(w http.ResponseWriter, r *http.Request) {}
	testSB.AddHandler(http.MethodHead, "/path", handlerHead)
	handleHead, _, _ := testSB.srv.handlers.Lookup(http.MethodHead, "/path")
	if handleHead == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerPut := func(w http.ResponseWriter, r *http.Request) {}
	testSB.AddHandler(http.MethodPut, "/path", handlerPut)
	handlePut, _, _ := testSB.srv.handlers.Lookup(http.MethodPut, "/path")
	if handlePut == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerPatch := func(w http.ResponseWriter, r *http.Request) {}
	testSB.AddHandler(http.MethodPatch, "/path", handlerPatch)
	handlePatch, _, _ := testSB.srv.handlers.Lookup(http.MethodPatch, "/path")
	if handlePatch == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerOptions := func(w http.ResponseWriter, r *http.Request) {}
	testSB.AddHandler(http.MethodOptions, "/path", handlerOptions)
	handleOptions, _, _ := testSB.srv.handlers.Lookup(http.MethodOptions, "/path")
	if handleOptions == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerDelete := func(w http.ResponseWriter, r *http.Request) {}
	testSB.AddHandler(http.MethodDelete, "/path", handlerDelete)
	handleDelete, _, _ := testSB.srv.handlers.Lookup(http.MethodDelete, "/path")
	if handleDelete == nil {
		t.Errorf("error: expected handle not nil")
	}
}

func TestAddGroup(t *testing.T) {
	testSB := Build(port)
	testSB.AddGroup("/test")
}

func TestAddGroupHandler(t *testing.T) {
	testSB := Build(port)
	gh := testSB.AddGroup("/test")

	handlerGet := func(w http.ResponseWriter, r *http.Request) {}
	gh.AddGroupHandler(http.MethodGet, "/path", handlerGet)
	handleGet, _, _ := gh.gr.server.handlers.Lookup(http.MethodGet, "/test/path")
	if handleGet == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerPost := func(w http.ResponseWriter, r *http.Request) {}
	gh.AddGroupHandler(http.MethodPost, "/path", handlerPost)
	handlePost, _, _ := gh.gr.server.handlers.Lookup(http.MethodPost, "/test/path")
	if handlePost == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerHead := func(w http.ResponseWriter, r *http.Request) {}
	gh.AddGroupHandler(http.MethodHead, "/path", handlerHead)
	handleHead, _, _ := gh.gr.server.handlers.Lookup(http.MethodHead, "/test/path")
	if handleHead == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerPut := func(w http.ResponseWriter, r *http.Request) {}
	gh.AddGroupHandler(http.MethodPut, "/path", handlerPut)
	handlePut, _, _ := gh.gr.server.handlers.Lookup(http.MethodPut, "/test/path")
	if handlePut == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerPatch := func(w http.ResponseWriter, r *http.Request) {}
	gh.AddGroupHandler(http.MethodPatch, "/path", handlerPatch)
	handlePatch, _, _ := gh.gr.server.handlers.Lookup(http.MethodPatch, "/test/path")
	if handlePatch == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerOptions := func(w http.ResponseWriter, r *http.Request) {}
	gh.AddGroupHandler(http.MethodOptions, "/path", handlerOptions)
	handleOptions, _, _ := gh.gr.server.handlers.Lookup(http.MethodOptions, "/test/path")
	if handleOptions == nil {
		t.Errorf("error: expected handle not nil")
	}

	handlerDelete := func(w http.ResponseWriter, r *http.Request) {}
	gh.AddGroupHandler(http.MethodDelete, "/path", handlerDelete)
	handleDelete, _, _ := gh.gr.server.handlers.Lookup(http.MethodDelete, "/test/path")
	if handleDelete == nil {
		t.Errorf("error: expected handle not nil")
	}

	testSB = gh.Return()
}

func TestAddFilesServer(t *testing.T) {
	testSB := Build(port)
	testSB.AddFilesServer("/*filepath", "/root")
	handle, _, _ := testSB.srv.handlers.Lookup(http.MethodGet, "/path*")
	if handle == nil {
		t.Errorf("error: expected handle not nil")
	}
}

func TestAddGroupFilesServer(t *testing.T) {
	testSB := Build(port)
	gb := testSB.AddGroup("/test")
	gb.AddGroupFilesServer("/*filepath", "/root")
	testSB = gb.Return()
	handle, _, _ := testSB.srv.handlers.Lookup(http.MethodGet, "/test/path*")
	if handle == nil {
		t.Errorf("error: expected handle not nil")
	}
}
