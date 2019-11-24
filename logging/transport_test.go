package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/izumin5210/hx"
	"github.com/izumin5210/hx/logging"
)

func TestWith(t *testing.T) {
	loc := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Date(2019, time.November, 24, 24, 32, 48, 0, loc)
	defer logging.SetNow(func() time.Time { return now })()
	respTime := 32 * time.Millisecond
	defer logging.SetSince(func(time.Time) time.Duration { return respTime })()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/ping":
			json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
		case r.Method == http.MethodGet && r.URL.Path == "/sleep":
			time.Sleep(100 * time.Millisecond)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		log := log.New(&buf, "", 0)

		err := hx.Get(context.Background(), ts.URL+"/ping",
			hx.TransportFunc(logging.With(log)),
			hx.WhenFailure(hx.AsError()),
		)

		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}

		if got, want := buf.String(),
			"Request HTTP/1.1 GET "+ts.URL+"/ping\n"+
				"Response 200 OK: GET "+ts.URL+"/ping (32ms)\n"; got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("failure", func(t *testing.T) {
		var buf bytes.Buffer
		log := log.New(&buf, "", 0)

		err := hx.Get(context.Background(), ts.URL+"/foobar",
			hx.TransportFunc(logging.With(log)),
			hx.WhenFailure(hx.AsError()),
		)

		if err == nil {
			t.Error("returned nil, want an error")
		}

		if got, want := buf.String(),
			"Request HTTP/1.1 GET "+ts.URL+"/foobar\n"+
				"Response 404 Not Found: GET "+ts.URL+"/foobar (32ms)\n"; got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("error", func(t *testing.T) {
		var buf bytes.Buffer
		log := log.New(&buf, "", 0)

		err := hx.Get(context.Background(), ts.URL+"/sleep",
			hx.TransportFunc(logging.With(log)),
			hx.WhenFailure(hx.AsError()),
			hx.Timeout(10*time.Millisecond),
		)

		if err == nil {
			t.Error("returned nil, want an error")
		}

		if got, want := buf.String(),
			"Request HTTP/1.1 GET "+ts.URL+"/sleep\n"+
				"Response error: net/http: request canceled: GET "+ts.URL+"/sleep (32ms)\n"; got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})
}
