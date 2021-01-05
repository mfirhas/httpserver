package httpserver

import (
	"net/http"
	"testing"
)

func TestUse(t *testing.T) {
	testM := func(next http.HandlerFunc, params ...interface{}) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {}
	}
	s := newServer()
	s.Use(testM)
	if len(s.middlewares) != 3 { // all newServer called add the middlewares
		t.Errorf("%s expected %d, returned %d", t.Name(), 2, len(s.middlewares))
	}
}

func TestServerChainMiddlewares(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {}
	srv := newServer()
	srv.chainMiddlewares(handler, TestMiddleware)
}

func TestGroupChainMiddlewares(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {}
	srv := newServer()
	grp := srv.Group("/test", TestMiddleware)
	grp.chainMiddlewares(handler, TestMiddleware)
}
