package httpserver

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_router "github.com/julienschmidt/httprouter"
	_cors "github.com/rs/cors"
)

type Builder interface {
	WithIdleTimeout(time.Duration) *ServerBuilder
	WithCors(*Cors) *ServerBuilder
	WithLogger() *ServerBuilder
	WithTLS(*tls.Config) *ServerBuilder
	WithPanicHandler(func(w http.ResponseWriter, r *http.Request, rcv ...interface{})) *ServerBuilder
	WithNotFoundHandler(http.HandlerFunc) *ServerBuilder
	WithMiddleware(Middleware) *ServerBuilder

	AddHandler(methodName string, path string, handler http.HandlerFunc, middlewares ...Middleware) *ServerBuilder
	AddFilesServer(filePath string, rootPath string, middlewares ...Middleware) *ServerBuilder

	Start() // blocking
}

type ServerBuilder struct {
	srv *Server
}

func Build(port uint16) *ServerBuilder {
	return &ServerBuilder{
		srv: &Server{
			port:     port,
			handlers: _router.New(),
			logger:   log.New(os.Stderr, "", 0),
		},
	}
}

func (sb *ServerBuilder) WithIdleTimeout(idleTimeout time.Duration) *ServerBuilder {
	sb.srv.idleTimeout = idleTimeout
	return sb
}

func (sb *ServerBuilder) WithCors(cors *Cors) *ServerBuilder {
	sb.srv.cors = _cors.New(_cors.Options{
		AllowedOrigins:     cors.AllowedOrigins,
		AllowedMethods:     cors.AllowedMethods,
		AllowedHeaders:     cors.AllowedHeaders,
		ExposedHeaders:     cors.ExposedHeaders,
		MaxAge:             cors.MaxAge,
		AllowCredentials:   cors.AllowCredentials,
		OptionsPassthrough: true,
		Debug:              cors.IsDebug,
	})
	return sb
}

func (sb *ServerBuilder) WithLogger(wr ...io.Writer) *ServerBuilder {
	buff := make(buffer, 10<<20)
	var w io.Writer = os.Stderr
	if len(wr) > 0 {
		if wr[0] != nil {
			w = wr[0]
		}
	}
	go write(buff, w)
	sb.srv.logger = log.New(buff, "", 0)
	sb.srv.middlewares = append(sb.srv.middlewares, sb.srv.log)
	return sb
}

func (sb *ServerBuilder) WithTLS(tls *tls.Config) *ServerBuilder {
	sb.srv.tls = tls
	return sb
}

func (sb *ServerBuilder) WithPanicHandler(panicHandler PanicHandler) *ServerBuilder {
	sb.srv.panicHandler = panicHandler
	return sb
}

func (sb *ServerBuilder) WithNotFoundHandler(notFoundHandlerFunc http.HandlerFunc) *ServerBuilder {
	var notFoundHandler http.Handler = &notFound{notFoundHandlerFunc}
	sb.srv.notFoundHandler = notFoundHandler
	return sb
}

func (sb *ServerBuilder) WithMiddleware(middleware Middleware) *ServerBuilder {
	sb.srv.middlewares = append(sb.srv.middlewares, middleware)
	return sb
}

func (sb *ServerBuilder) AddHandler(methodName string, path string, handler http.HandlerFunc, middlewares ...Middleware) *ServerBuilder {
	switch methodName {
	case http.MethodGet:
		sb.srv.GET(path, handler, middlewares...)
		break
	case http.MethodHead:
		sb.srv.HEAD(path, handler, middlewares...)
		break
	case http.MethodPost:
		sb.srv.POST(path, handler, middlewares...)
		break
	case http.MethodPut:
		sb.srv.PUT(path, handler, middlewares...)
		break
	case http.MethodDelete:
		sb.srv.DELETE(path, handler, middlewares...)
		break
	case http.MethodPatch:
		sb.srv.PATCH(path, handler, middlewares...)
		break
	case http.MethodOptions:
		sb.srv.OPTIONS(path, handler, middlewares...)
		break
	default:
		panic("httpserver: ServerBuilder.AddHandler method name is not identified!")
	}
	return sb
}

func (sb *ServerBuilder) AddFilesServer(filePath string, rootPath string, middlewares ...Middleware) *ServerBuilder {
	sb.srv.FILES(filePath, rootPath, middlewares...)
	return sb
}

func (sb *ServerBuilder) Run() {
	sb.srv.Run()
}

type GroupBuilder struct {
	sb *ServerBuilder
	gr *Group
}

func (sb *ServerBuilder) AddGroup(prefix string, middlewares ...Middleware) *GroupBuilder {
	return &GroupBuilder{
		sb: sb,
		gr: &Group{
			server:      sb.srv,
			prefix:      prefix,
			middlewares: middlewares,
		},
	}
}

func (gb *GroupBuilder) AddGroupHandler(methodName string, path string, handler http.HandlerFunc, middlewares ...Middleware) *GroupBuilder {
	switch methodName {
	case http.MethodGet:
		gb.gr.GET(path, handler, middlewares...)
		break
	case http.MethodHead:
		gb.gr.HEAD(path, handler, middlewares...)
		break
	case http.MethodPost:
		gb.gr.POST(path, handler, middlewares...)
		break
	case http.MethodPut:
		gb.gr.PUT(path, handler, middlewares...)
		break
	case http.MethodDelete:
		gb.gr.DELETE(path, handler, middlewares...)
		break
	case http.MethodPatch:
		gb.gr.PATCH(path, handler, middlewares...)
		break
	case http.MethodOptions:
		gb.gr.OPTIONS(path, handler, middlewares...)
		break
	default:
		panic("httpserver: GroupBuilder.AddHandler method name is not identified!")
	}
	return gb
}

func (gb *GroupBuilder) AddGroupFilesServer(filePath string, rootPath string, middlewares ...Middleware) *GroupBuilder {
	gb.gr.FILES(filePath, rootPath, middlewares...)
	return gb
}

func (gb *GroupBuilder) Return() *ServerBuilder {
	gb.sb.srv = gb.gr.server
	gb.gr = nil
	return gb.sb
}
