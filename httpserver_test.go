package httpserver

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	_router "github.com/julienschmidt/httprouter"
)

var (
	TestMiddleware = func(next http.HandlerFunc, params ...interface{}) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {}
	}
)

func newServer() *Server {
	w := make(buffer, 10)
	go write(w)
	cors := &Cors{}
	srv := New(&Opts{
		Port:            8080,
		Cors:            cors,
		EnableLogger:    true,
		PanicHandler:    func(w http.ResponseWriter, r *http.Request, rcv ...interface{}) {},
		NotFoundHandler: func(w http.ResponseWriter, r *http.Request) {},
	})
	srv.Use(TestMiddleware)
	return srv
}

func TestNew(t *testing.T) {
	newServer()
}

func TestListenError(t *testing.T) {
	srv := newServer()
	go func() {
		err := <-srv.ListenError()
		if err == nil {
			t.Errorf("%s expected non-empty error, return null", t.Name())
		}
		if err.Error() != "error" {
			t.Errorf("%s expected %s, returned %s", t.Name(), "error", err.Error())
		}
	}()
	srv.errChan <- fmt.Errorf("error")
}

func TestWriteHeader(t *testing.T) {
	w := &httptest.ResponseRecorder{}
	rw := &responseWriter{ResponseWriter: w}
	rw.WriteHeader(200)
}

func TestF_ReqIDEmpty(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {}
	var ps _router.Params
	w := &httptest.ResponseRecorder{}
	r, _ := http.NewRequest("GET", "/health-check", nil)
	thisF := f(next)
	thisF(w, r, ps)
	if r.Header.Get("Request-Id") == "" {
		t.Errorf("%s expected Header Request-Id not empty, found empty", t.Name())
	}
}

func TestF_ReqIDEmpty_XReqIDNotEmpty(t *testing.T) {
	testXRequestID := "testXRequestID"
	next := func(w http.ResponseWriter, r *http.Request) {}
	var ps _router.Params
	w := &httptest.ResponseRecorder{}
	r, _ := http.NewRequest("GET", "/health-check", nil)
	r.Header.Set("X-Request-Id", testXRequestID)
	thisF := f(next)
	thisF(w, r, ps)
	if r.Header.Get("Request-Id") == "" {
		t.Errorf("%s expected Header Request-Id not empty, found empty", t.Name())
	}
}

func TestF_WithPathParams(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {}
	var ps _router.Params
	ps = append(ps, _router.Param{
		Key:   "key",
		Value: "value",
	})
	w := &httptest.ResponseRecorder{}
	r, _ := http.NewRequest("GET", "/health-check", nil)
	thisF := f(next)
	thisF(w, r, ps)
	if r.Header.Get("Request-Id") == "" {
		t.Errorf("%s expected Header Request-Id not empty, found empty", t.Name())
	}
}

func TestRecoverPanic(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		testPanic := []int{1, 2}
		_ = testPanic[2]
	}
	w := &httptest.ResponseRecorder{}
	r, _ := http.NewRequest("GET", "/health-check", nil)
	srv := newServer()
	h := srv.recoverPanic(next)
	h(w, r)
}

func TestResponseHeader(t *testing.T) {
	w := &httptest.ResponseRecorder{}
	rw := &responseWriter{w, 200, "", ""}
	responseHeader(rw, 200)
}

func TestResponse(t *testing.T) {
	w := &httptest.ResponseRecorder{}
	Response(w, 200, []byte("test"))
}

func TestResponseJSON(t *testing.T) {
	w := &httptest.ResponseRecorder{}
	if err := ResponseJSON(w, 200, []byte("test")); err != nil {
		t.Errorf("%s expected null error, found not null", t.Name())
	}
}

func TestResponseString(t *testing.T) {
	w := &httptest.ResponseRecorder{}
	ResponseString(w, 200, "string")
}

var testSrv = newServer()

func TestGET(t *testing.T) {
	testSrv.GET("/get", testHandler, TestMiddleware)
}

func TestHEAD(t *testing.T) {
	testSrv.HEAD("/head", testHandler, TestMiddleware)
}

func TestPOST(t *testing.T) {
	testSrv.POST("/post", testHandler, TestMiddleware)
}

func TestPUT(t *testing.T) {
	testSrv.PUT("/put", testHandler, TestMiddleware)
}

func TestDELETE(t *testing.T) {
	testSrv.DELETE("/delete", testHandler, TestMiddleware)
}

func TestPATCH(t *testing.T) {
	testSrv.PATCH("/patch", testHandler, TestMiddleware)
}

func TestOPTIONS(t *testing.T) {
	testSrv.OPTIONS("/options", testHandler, TestMiddleware)
}

func TestFILES_OnSuccess(t *testing.T) {
	testSrv.FILES("/test/*filepath", "/test/")
	go testSrv.Run()
	go testSrv.Run()
}

func TestFILES_OnFailed(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("expected")
		}
	}()
	testSrv.FILES("/test/*filepat", "/test/")
}

func TestRenderHTML(t *testing.T) {
	tmplName := "test"
	tmpl := `<h1>Test</h1>`
	funcMap := template.FuncMap{
		"f1": func() template.HTML {
			return ""
		},
	}
	RenderHTML(tmplName, tmpl, nil, funcMap)
}

func TestResponseHTML(t *testing.T) {
	tmplName := "test"
	tmpl := `<h1>Test</h1>`
	funcMap := template.FuncMap{
		"f1": func() template.HTML {
			return ""
		},
	}
	w := httptest.NewRecorder()
	ResponseHTML(w, tmplName, tmpl, nil, funcMap)
}

func TestRenderMultiHTML(t *testing.T) {
	tmplName := "test"
	tmplNameToTmpl := map[string]string{
		"test":  "<h1>test</h1>",
		"test2": "<h2>test</h2>",
	}
	funcMap := template.FuncMap{
		"f1": func() template.HTML {
			return ""
		},
	}
	RenderMultiHTML(tmplName, tmplNameToTmpl, nil, funcMap)
}

func TestResponseMultiHTML(t *testing.T) {
	tmplName := "test"
	tmplNameToTmpl := map[string]string{
		"test":  "<h1>test</h1>",
		"test2": "<h2>test</h2>",
	}
	funcMap := template.FuncMap{
		"f1": func() template.HTML {
			return ""
		},
	}
	w := httptest.NewRecorder()
	ResponseMultiHTML(w, tmplName, tmplNameToTmpl, nil, funcMap)
}

func TestLoadTemplate(t *testing.T) {
	LoadTemplate("_")
}
