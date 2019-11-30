package hxzap_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/izumin5210/hx"
	"github.com/izumin5210/hx/plugins/hxzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestWith(t *testing.T) {
	loc := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Date(2019, time.November, 24, 24, 32, 48, 0, loc)
	defer hxzap.SetNow(func() time.Time { return now })()
	respTime := 32 * time.Millisecond
	defer hxzap.SetSince(func(time.Time) time.Duration { return respTime })()

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
		core, logs := observer.New(zapcore.DebugLevel)

		err := hx.Get(context.Background(), ts.URL+"/ping",
			hx.TransportFunc(hxzap.With(zap.New(core))),
			hx.WhenFailure(hx.AsError()),
		)

		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}

		if got, want := logs.Len(), 2; got != want {
			t.Errorf("logged %d items, want %d", got, want)
		} else {
			if got, want := logs.All()[0].ContextMap(), map[string]interface{}{
				"proto":          "HTTP/1.1",
				"method":         "GET",
				"host":           strings.TrimPrefix(ts.URL, "http://"),
				"path":           "/ping",
				"url":            ts.URL + "/ping",
				"content_length": int64(0),
			}; !reflect.DeepEqual(got, want) {
				t.Errorf("got:\n%v\nwant:\n%v", got, want)
			}
			if got, want := logs.All()[1].ContextMap(), map[string]interface{}{
				"proto":          "HTTP/1.1",
				"method":         "GET",
				"host":           strings.TrimPrefix(ts.URL, "http://"),
				"path":           "/ping",
				"url":            ts.URL + "/ping",
				"status":         "200 OK",
				"status_code":    int64(200),
				"response_time":  respTime,
				"content_length": int64(19),
			}; !reflect.DeepEqual(got, want) {
				t.Errorf("got:\n%v\nwant:\n%v", got, want)
			}
		}
	})

	t.Run("failure", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)

		err := hx.Get(context.Background(), ts.URL+"/foobar",
			hx.TransportFunc(hxzap.With(zap.New(core))),
			hx.WhenFailure(hx.AsError()),
		)

		if err == nil {
			t.Error("returned nil, want an error")
		}

		if got, want := logs.Len(), 2; got != want {
			t.Errorf("logged %d items, want %d", got, want)
		} else {
			if got, want := logs.All()[0].ContextMap(), map[string]interface{}{
				"proto":          "HTTP/1.1",
				"method":         "GET",
				"host":           strings.TrimPrefix(ts.URL, "http://"),
				"path":           "/foobar",
				"url":            ts.URL + "/foobar",
				"content_length": int64(0),
			}; !reflect.DeepEqual(got, want) {
				t.Errorf("got:\n%v\nwant:\n%v", got, want)
			}
			if got, want := logs.All()[1].ContextMap(), map[string]interface{}{
				"proto":          "HTTP/1.1",
				"method":         "GET",
				"host":           strings.TrimPrefix(ts.URL, "http://"),
				"path":           "/foobar",
				"url":            ts.URL + "/foobar",
				"status":         "404 Not Found",
				"status_code":    int64(404),
				"response_time":  respTime,
				"content_length": int64(0),
			}; !reflect.DeepEqual(got, want) {
				t.Errorf("got:\n%v\nwant:\n%v", got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)

		err := hx.Get(context.Background(), ts.URL+"/sleep",
			hx.TransportFunc(hxzap.With(zap.New(core))),
			hx.WhenFailure(hx.AsError()),
			hx.Timeout(10*time.Millisecond),
		)

		if err == nil {
			t.Error("returned nil, want an error")
		}

		if got, want := logs.Len(), 2; got != want {
			t.Errorf("logged %d items, want %d", got, want)
		} else {
			if got, want := logs.All()[0].ContextMap(), map[string]interface{}{
				"proto":          "HTTP/1.1",
				"method":         "GET",
				"host":           strings.TrimPrefix(ts.URL, "http://"),
				"path":           "/sleep",
				"url":            ts.URL + "/sleep",
				"content_length": int64(0),
			}; !reflect.DeepEqual(got, want) {
				t.Errorf("got:\n%v\nwant:\n%v", got, want)
			}
			if got, want := logs.All()[1].ContextMap(), map[string]interface{}{
				"proto":         "HTTP/1.1",
				"method":        "GET",
				"host":          strings.TrimPrefix(ts.URL, "http://"),
				"path":          "/sleep",
				"url":           ts.URL + "/sleep",
				"error":         "net/http: request canceled",
				"response_time": respTime,
			}; !reflect.DeepEqual(got, want) {
				t.Errorf("got:\n%#v\nwant:\n%#v", got, want)
			}
		}
	})
}
