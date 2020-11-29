package httpserver

import (
	"net/http"
)

func (s *Server) Use(m ...Middleware) {
	for _, v := range m {
		s.middlewares = append(s.middlewares, v)
	}
}

// chainMiddlewares chain all middlewares to handler
func (s *Server) chainMiddlewares(handler http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	h := handler
	var ch http.HandlerFunc

	if len(middlewares) > 0 {
		lM := len(middlewares) - 1
		for i := lM; i >= 0; i-- {
			ch = middlewares[i](h)
			h = ch
		}
	}

	if len(s.middlewares) > 0 {
		lS := len(s.middlewares) - 1
		for i := lS; i >= 0; i-- {
			ch = s.middlewares[i](h)
			h = ch
		}
	}

	return h
}

func (g *Group) chainMiddlewares(handler http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	h := handler
	var ch http.HandlerFunc

	if len(middlewares) > 0 {
		lM := len(middlewares) - 1
		for i := lM; i >= 0; i-- {
			ch = middlewares[i](h)
			h = ch
		}
	}

	if len(g.middlewares) > 0 {
		lG := len(g.middlewares) - 1
		for i := lG; i >= 0; i-- {
			ch = g.middlewares[i](h)
			h = ch
		}
	}

	if len(g.server.middlewares) > 0 {
		lS := len(g.server.middlewares) - 1
		for i := lS; i >= 0; i-- {
			ch = g.server.middlewares[i](h)
			h = ch
		}
	}

	return h
}
