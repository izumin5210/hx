package hxutil

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDrainResponseBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/ping":
			w.Write([]byte("pong"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	t.Run("success", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/ping")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()

		err = DrainResponseBody(resp)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		} else if got, want := string(data), "pong"; got != want {
			t.Errorf("returned %q, want %q", got, want)
		}
	})

	t.Run("failure", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/ping")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp.Body.Close()

		err = DrainResponseBody(resp)
		if err == nil {
			t.Errorf("returned nil, want an error")
		}
	})
}
