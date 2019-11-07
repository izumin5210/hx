package hx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/izumin5210/hx"
)

func TestClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/ping":
			r.Write(bytes.NewBufferString("pong"))
		case r.Method == http.MethodGet && r.URL.Path == "/echo":
			msg := r.URL.Query().Get("message")
			if msg == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err := json.NewEncoder(w).Encode(map[string]string{"message": msg})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case r.Method == http.MethodPost && r.URL.Path == "/echo":
			out := make(map[string]interface{})
			err := json.NewDecoder(r.Body).Decode(&out)
			if err != nil || out["message"] == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			err = json.NewEncoder(w).Encode(out)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case r.Method == http.MethodGet && r.URL.Path == "/basic_auth":
			if user, pass, ok := r.BasicAuth(); !(ok && user == "foo" && pass == "bar") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		case r.Method == http.MethodGet && r.URL.Path == "/bearer_auth":
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token != "tokentoken" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		case r.Method == http.MethodGet && r.URL.Path == "/error":
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "invalid argument"})
		case r.Method == http.MethodGet && r.URL.Path == "/timeout":
			time.Sleep(1 * time.Second)
			err := json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	checkStatusFromError := func(t *testing.T, err error, st int) {
		t.Helper()
		if err == nil {
			t.Error("returned nil, want an error")
		} else if reqErr, ok := err.(*hx.ResponseError); !ok {
			t.Errorf("returned %v, want *hx.ResponseError", err)
		} else if reqErr.Response == nil {
			t.Error("returned error has no response")
		} else if got, want := reqErr.Response.StatusCode, st; got != want {
			t.Errorf("returned status code is %d, want %d", got, want)
		}
	}
	checkErrorIsWrapped := func(t *testing.T, err error) {
		t.Helper()
		if err == nil {
			t.Error("returned nil, want an error")
		} else if reqErr, ok := err.(*hx.ResponseError); !ok {
			t.Errorf("returned %v, want *hx.ResponseError", err)
		} else if reqErr.Unwrap() == reqErr {
			t.Error("returned error wrapped no errors")
		}
	}
	checkErrorIsNotWrapped := func(t *testing.T, err error) {
		t.Helper()
		if err == nil {
			t.Error("returned nil, want an error")
		} else if reqErr, ok := err.(*hx.ResponseError); !ok {
			t.Errorf("returned %v, want *hx.ResponseError", err)
		} else if reqErr.Unwrap() != reqErr {
			t.Errorf("returned error wrapped %v, want nil", reqErr.Unwrap())
		}
	}

	defer ts.Close()

	t.Run("simple", func(t *testing.T) {
		err := hx.Get(context.Background(), ts.URL+"/ping")
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
	})

	t.Run("receive json", func(t *testing.T) {
		var out struct {
			Message string `json:"message"`
		}
		err := hx.Get(context.Background(), ts.URL+"/echo",
			hx.Query("message", "It, Works!"),
			hx.WhenSuccess(hx.AsJSON(&out)),
		)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
		if got, want := out.Message, "It, Works!"; got != want {
			t.Errorf("returned %q, want %q", got, want)
		}
	})

	t.Run("when error", func(t *testing.T) {
		t.Run("ignore", func(t *testing.T) {
			var out struct {
				Message string `json:"message"`
			}
			err := hx.Get(context.Background(), ts.URL+"/echo",
				hx.WhenSuccess(hx.AsJSON(&out)),
			)
			if err != nil {
				t.Errorf("returned %v, want nil", err)
			}
			if got, want := out.Message, ""; got != want {
				t.Errorf("returned %q, want %q", got, want)
			}
		})

		t.Run("handle", func(t *testing.T) {
			var out struct {
				Message string `json:"message"`
			}
			err := hx.Get(context.Background(), ts.URL+"/echo",
				hx.WhenSuccess(hx.AsJSON(&out)),
				hx.WhenFailure(hx.AsError()),
			)
			checkStatusFromError(t, err, http.StatusBadRequest)
			checkErrorIsNotWrapped(t, err)
		})

		t.Run("failed to decode response", func(t *testing.T) {
			var out struct {
				Message string `json:"message"`
			}
			err := hx.Get(context.Background(), ts.URL+"/ping",
				hx.WhenSuccess(hx.AsJSON(&out)),
				hx.WhenFailure(hx.AsError()),
			)
			checkStatusFromError(t, err, http.StatusOK)
			checkErrorIsWrapped(t, err)
		})

		t.Run("AsErrorOf", func(t *testing.T) {
			err := hx.Get(context.Background(), ts.URL+"/error",
				hx.WhenStatus(hx.AsErrorOf(&fakeError{}), http.StatusBadRequest),
				hx.WhenFailure(hx.AsError()),
			)
			if err == nil {
				t.Error("returned nil, want an error")
			} else if reqErr, ok := err.(*hx.ResponseError); !ok {
				t.Errorf("returned %v, want *hx.ResponseError", err)
			} else if rawErr := reqErr.Err; rawErr == nil {
				fmt.Println(reqErr)
				t.Error("returned error wrapped no errors")
			} else if fakeErr, ok := rawErr.(*fakeError); !ok {
				t.Errorf("wrapped error is unknown: %v", rawErr)
			} else if got, want := fakeErr.Message, "invalid argument"; got != want {
				t.Errorf("wrapped error has message %v, want %v", got, want)
			}
		})
	})

	t.Run("With BaseURL", func(t *testing.T) {
		u, _ := url.Parse(ts.URL)
		cli := hx.NewClient(hx.BaseURL(u))
		err := cli.Get(context.Background(), "/ping",
			hx.WhenFailure(hx.AsError()),
		)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
	})

	t.Run("With BasicAuth", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			err := hx.Get(context.Background(), ts.URL+"/basic_auth",
				hx.BasicAuth("foo", "bar"),
				hx.WhenFailure(hx.AsError()),
			)
			if err != nil {
				t.Errorf("returned %v, want nil", err)
			}
		})

		t.Run("failure", func(t *testing.T) {
			err := hx.Get(context.Background(), ts.URL+"/basic_auth",
				hx.BasicAuth("baz", "qux"),
				hx.WhenFailure(hx.AsError()),
			)
			checkStatusFromError(t, err, http.StatusUnauthorized)
			checkErrorIsNotWrapped(t, err)
		})
	})

	t.Run("with Bearer", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			err := hx.Get(context.Background(), ts.URL+"/bearer_auth",
				hx.Bearer("tokentoken"),
				hx.WhenFailure(hx.AsError()),
			)
			if err != nil {
				t.Errorf("returned %v, want nil", err)
			}
		})

		t.Run("failure", func(t *testing.T) {
			err := hx.Get(context.Background(), ts.URL+"/bearer_auth",
				hx.Bearer("tokentokentoken"),
				hx.WhenFailure(hx.AsError()),
			)
			checkStatusFromError(t, err, http.StatusUnauthorized)
			checkErrorIsNotWrapped(t, err)
		})
	})

	t.Run("with Timeout", func(t *testing.T) {
		var out struct {
			Message string `json:"message"`
		}
		err := hx.Get(context.Background(), ts.URL+"/timeout",
			hx.WhenSuccess(hx.AsJSON(&out)),
			hx.Timeout(10*time.Millisecond),
			hx.WhenFailure(hx.AsError()),
		)
		if err == nil {
			t.Error("returned nil, want an error")
		}
	})

	t.Run("with Client", func(t *testing.T) {
		cli := &http.Client{
			Timeout: 10 * time.Millisecond,
		}
		err := hx.Get(context.Background(), ts.URL+"/timeout",
			hx.HTTPClient(cli),
			hx.WhenFailure(hx.AsError()),
		)
		if err == nil {
			t.Error("returned nil, want an error")
		}
	})

	t.Run("with Transport", func(t *testing.T) {
		transport := &fakeTransport{
			RoundTripFunc: func(rt http.RoundTripper, req *http.Request) (*http.Response, error) {
				req.SetBasicAuth("foo", "bar")
				return rt.RoundTrip(req)
			},
		}
		err := hx.Get(context.Background(), ts.URL+"/basic_auth",
			hx.Transport(transport),
			hx.WhenFailure(hx.AsError()),
		)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
	})

	t.Run("with TransportFrom", func(t *testing.T) {
		err := hx.Get(context.Background(), ts.URL+"/basic_auth",
			hx.TransportFrom(func(base http.RoundTripper) http.RoundTripper {
				return &fakeTransport{
					Base: base,
					RoundTripFunc: func(rt http.RoundTripper, req *http.Request) (*http.Response, error) {
						req.SetBasicAuth("foo", "bar")
						return rt.RoundTrip(req)
					},
				}
			}),
			hx.WhenFailure(hx.AsError()),
		)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
	})
}

type fakeError struct {
	Message string `json:"message"`
}

func (e fakeError) Error() string { return e.Message }

type fakeTransport struct {
	Base          http.RoundTripper
	RoundTripFunc func(http.RoundTripper, *http.Request) (*http.Response, error)
}

func (t *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return t.RoundTripFunc(base, req)
}
