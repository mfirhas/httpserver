package httpserver

import (
	"net/http"
	"testing"
)

var (
	groupServer = newServer()
	group       = groupServer.Group("/test", TestMiddleware)
	testHandler = func(w http.ResponseWriter, r *http.Request) {}
)

func TestGroupGET(t *testing.T) {
	group.GET("/get", testHandler, TestMiddleware)
}

func TestGroupHEAD(t *testing.T) {
	group.HEAD("/head", testHandler, TestMiddleware)
}

func TestGroupPOST(t *testing.T) {
	group.POST("/post", testHandler, TestMiddleware)
}

func TestGroupPUT(t *testing.T) {
	group.PUT("/put", testHandler, TestMiddleware)
}

func TestGroupDELETE(t *testing.T) {
	group.DELETE("/delete", testHandler, TestMiddleware)
}

func TestGroupPATCH(t *testing.T) {
	group.PATCH("/patch", testHandler, TestMiddleware)
}

func TestGroupOPTIONS(t *testing.T) {
	group.OPTIONS("/options", testHandler, TestMiddleware)
}

func TestGroupFILES(t *testing.T) {
	group.FILES("/test/*filepath", "/test/")
}
