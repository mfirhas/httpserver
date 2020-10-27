package main

import (
	"fmt"
	"net/http"
	"os"

	_httpserver "github.com/mfathirirhas/httpserver"
)

func main() {
	srv := _httpserver.New(&_httpserver.Opts{
		Port:         8080,
		EnableLogger: true,
		IdleTimeout:  10,
	})

	m1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("This middleware m1 : ", os.Getpid())
			next(w, r)
		}
	}
	srv.GET("/handler1", Handler1, m1)
	srv.POST("/handler2", Handler2)
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
