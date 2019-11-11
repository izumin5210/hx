package hx_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/izumin5210/hx"
)

func TestResponseHandlerCond(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/ping":
			status, _ := strconv.Atoi(r.URL.Query().Get("status"))
			if status == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(status)
			w.Write([]byte("pong"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	defer l.Close()
	freePort := l.Addr().(*net.TCPAddr).Port

	ctx := context.Background()

	t.Run("client error", func(t *testing.T) {
		for _, st := range []int{400, 401, 498, 499} {
			t.Run(fmt.Sprint(st), func(t *testing.T) {
				err := hx.Get(ctx, ts.URL+"/ping",
					hx.Query("status", fmt.Sprint(st)),
					hx.WhenClientError(hx.AsError()),
				)
				if err == nil {
					t.Error("returned nil, want an error")
				}
			})
		}
		for _, st := range []int{398, 399, 500, 501} {
			t.Run(fmt.Sprint(st), func(t *testing.T) {
				err := hx.Get(ctx, ts.URL+"/ping",
					hx.Query("status", fmt.Sprint(st)),
					hx.WhenClientError(hx.AsError()),
				)
				if err != nil {
					t.Errorf("returned %v, want nil", err)
				}
			})
		}
	})

	t.Run("server error", func(t *testing.T) {
		for _, st := range []int{500, 501, 598, 599} {
			t.Run(fmt.Sprint(st), func(t *testing.T) {
				err := hx.Get(ctx, ts.URL+"/ping",
					hx.Query("status", fmt.Sprint(st)),
					hx.WhenServerError(hx.AsError()),
				)
				if err == nil {
					t.Error("returned nil, want an error")
				}
			})
		}
		for _, st := range []int{498, 499, 600, 601} {
			t.Run(fmt.Sprint(st), func(t *testing.T) {
				err := hx.Get(ctx, ts.URL+"/ping",
					hx.Query("status", fmt.Sprint(st)),
					hx.WhenServerError(hx.AsError()),
				)
				if err != nil {
					t.Errorf("returned %v, want nil", err)
				}
			})
		}
	})

	t.Run("network error", func(t *testing.T) {
		for _, st := range []int{400, 500} {
			t.Run(fmt.Sprint(st), func(t *testing.T) {
				err := hx.Get(ctx, ts.URL+"/ping",
					hx.Timeout(10*time.Millisecond),
					hx.Query("status", fmt.Sprint(st)),
					hx.When(hx.IsRoundTripError(), hx.AsError()),
				)
				if err != nil {
					t.Errorf("returned %v, want nil", err)
				}
			})
		}
		t.Run("when server stopped", func(t *testing.T) {
			err := hx.Get(ctx, fmt.Sprintf("http://localhost:%d/ping", freePort),
				hx.Timeout(10*time.Millisecond),
				hx.When(hx.IsRoundTripError(), hx.AsError()),
			)
			if err == nil {
				t.Error("returned nil, want an error")
			}
		})
	})

	t.Run("Any", func(t *testing.T) {
		cond := hx.Any(hx.IsServerError(), hx.IsRoundTripError())
		t.Run(fmt.Sprint(500), func(t *testing.T) {
			err := hx.Get(ctx, ts.URL+"/ping",
				hx.Timeout(10*time.Millisecond),
				hx.Query("status", fmt.Sprint(500)),
				hx.When(cond, hx.AsError()),
			)
			if err == nil {
				t.Error("returned nil, want an error")
			}
		})
		t.Run(fmt.Sprint(400), func(t *testing.T) {
			err := hx.Get(ctx, ts.URL+"/ping",
				hx.Timeout(10*time.Millisecond),
				hx.Query("status", fmt.Sprint(400)),
				hx.When(cond, hx.AsError()),
			)
			if err != nil {
				t.Errorf("returned %v, want nil", err)
			}
		})
		t.Run("when server stopped", func(t *testing.T) {
			err := hx.Get(ctx, fmt.Sprintf("http://localhost:%d/ping", freePort),
				hx.Timeout(10*time.Millisecond),
				hx.When(cond, hx.AsError()),
			)
			if err == nil {
				t.Error("returned nil, want an error")
			}
		})
	})
}
