package hx_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/izumin5210/hx"
)

func TestInterceptor(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/ping":
			w.Write([]byte("pong"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	var buf bytes.Buffer

	err := hx.Get(context.Background(), ts.URL+"/ping",
		hx.WhenSuccess(hx.AsBytesBuffer(&buf)),
		hx.WhenFailure(hx.AsError()),
		hx.InterceptFunc(func(c *http.Client, req *http.Request, f hx.RequestFunc) (*http.Response, error) {
			buf.WriteString("1")
			resp, err := f(c, req)
			buf.WriteString("4")
			return resp, err
		}),
		hx.InterceptFunc(func(c *http.Client, req *http.Request, f hx.RequestFunc) (*http.Response, error) {
			buf.WriteString("2")
			resp, err := f(c, req)
			buf.WriteString("3")
			return resp, err
		}),
	)
	if err != nil {
		t.Errorf("returned %v, want nil", err)
	}

	if got, want := buf.String(), "12pong34"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
