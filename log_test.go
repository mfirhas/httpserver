package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrite(t *testing.T) {
	w := make(buffer, 5)
	b := []byte("test")
	if _, err := w.Write(b); err != nil {
		t.Errorf("%s failed.", t.Name())
	}
}

func TestWriter(t *testing.T) {
	w := make(buffer, 1)
	b := []byte("test")
	w <- b
	go write(w)
}

func TestLog(t *testing.T) {
	s := newServer()
	next := func(w http.ResponseWriter, r *http.Request) {}
	h := s.log(next)

	r, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Errorf("%s failed", t.Name())
	}
	w := httptest.NewRecorder()
	h(w, r)
	rw := newResponseWriter(w, "", "")
	h(rw, r)
}
