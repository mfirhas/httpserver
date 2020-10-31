package main

import (
	"fmt"
	"net/http"
	"os"

	_httpserver "github.com/mfathirirhas/httpserver"
)

// /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.key
// /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.csr
// /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.crt
// openssl req  -new  -newkey rsa:2048  -nodes  -keyout /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.key  -out /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.csr
// openssl  x509  -req  -days 365  -in /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.csr  -signkey /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.key  -out /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.crt

func main() {
	srv := _httpserver.New(&_httpserver.Opts{
		Port:         8080,
		EnableLogger: true,
		IdleTimeout:  10,
	})
	// srv.TLSConfig(
	// 	"/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.crt",
	// 	"/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.key",
	// )
	msrv := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("This is middleware for all handlers in this server")
			next(w, r)
		}
	}
	srv.Use(msrv)
	m1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("This middleware m1 : ", os.Getpid())
			next(w, r)
		}
	}
	srv.GET("/handler1", Handler1, m1)
	srv.POST("/handler2", Handler2)
	mg1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("This is group 1 middleware: ", os.Getpid())
			next(w, r)
		}
	}
	g1 := srv.Group("/v1", mg1)
	g1.GET("/handler1", Handler1)
	gh2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("This is group 1 handler 2 middleware: ", os.Getpid())
			next(w, r)
		}
	}
	g1.POST("/handler2", Handler2, gh2)
	srv.Run()
}

func Handler1(w http.ResponseWriter, r *http.Request) {
	allHeaders := r.Header
	allParams := r.URL.RawQuery
	_httpserver.ResponseString(w, r, http.StatusOK, fmt.Sprintf("Handler1: %s | %s", allHeaders, allParams))
}

func Handler2(w http.ResponseWriter, r *http.Request) {
	allHeaders := r.Header
	allParams := r.URL.RawQuery
	r.ParseForm()
	b1 := r.FormValue("key1")
	b2 := r.FormValue("key2")
	resp := make(map[string]interface{})
	resp["headers"] = allHeaders
	resp["params"] = allParams
	resp["body1"] = b1
	resp["body2"] = b2
	_httpserver.ResponseJSON(w, r, http.StatusOK, resp)
}
