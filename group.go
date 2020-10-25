package httpserver

import (
	"fmt"
	"net/http"
)

type Group struct {
	server *Server
	prefix string
}

func (s *Server) Group(prefix string) *Group {
	return &Group{
		server: s,
		prefix: prefix,
	}
}

func (g *Group) GET(path string, handler http.HandlerFunc) {
	g.server.handlers.GET(fmt.Sprintf("%s%s", g.prefix, path), f(g.server.recoverPanic(g.server.log(handler))))
}

func (g *Group) HEAD(path string, handler http.HandlerFunc) {
	g.server.handlers.HEAD(fmt.Sprintf("%s%s", g.prefix, path), f(g.server.recoverPanic(g.server.log(handler))))
}

func (g *Group) HEADGET(path string, handler http.HandlerFunc) {
	g.server.handlers.HEAD(fmt.Sprintf("%s%s", g.prefix, path), f(g.server.recoverPanic(g.server.log(handler))))
	g.server.handlers.GET(fmt.Sprintf("%s%s", g.prefix, path), f(g.server.recoverPanic(g.server.log(handler))))
}

func (g *Group) POST(path string, handler http.HandlerFunc) {
	g.server.handlers.POST(fmt.Sprintf("%s%s", g.prefix, path), f(g.server.recoverPanic(g.server.log(handler))))
}

func (g *Group) PUT(path string, handler http.HandlerFunc) {
	g.server.handlers.POST(fmt.Sprintf("%s%s", g.prefix, path), f(g.server.recoverPanic(g.server.log(handler))))
}

func (g *Group) DELETE(path string, handler http.HandlerFunc) {
	g.server.handlers.DELETE(fmt.Sprintf("%s%s", g.prefix, path), f(g.server.recoverPanic(g.server.log(handler))))
}

func (g *Group) PATCH(path string, handler http.HandlerFunc) {
	g.server.handlers.PATCH(fmt.Sprintf("%s%s", g.prefix, path), f(g.server.recoverPanic(g.server.log(handler))))
}

func (g *Group) OPTIONS(path string, handler http.HandlerFunc) {
	g.server.handlers.OPTIONS(fmt.Sprintf("%s%s", g.prefix, path), f(g.server.recoverPanic(g.server.log(handler))))
}
