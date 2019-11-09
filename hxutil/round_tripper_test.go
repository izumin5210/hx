package hxutil_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/izumin5210/hx/hxutil"
)

func TestRoundTripperFunc(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/echo":
			cnt, _ := strconv.Atoi(r.Header.Get("Count"))
			if cnt == 0 {
				cnt = 1
			}
			var buf bytes.Buffer
			io.Copy(&buf, r.Body)
			w.Write([]byte(strings.Repeat(buf.String(), cnt)))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cases := []struct {
		test string
		base http.RoundTripper
	}{
		{test: "no base"},
		{test: "specify base", base: http.DefaultTransport},
	}

	for _, tc := range cases {
		t.Run(tc.test, func(t *testing.T) {
			cli := &http.Client{
				Transport: hxutil.RoundTripperFunc(func(r *http.Request, rt http.RoundTripper) (*http.Response, error) {
					r.Header.Set("Count", "3")
					return rt.RoundTrip(r)
				}).Wrap(tc.base),
			}

			req, err := http.NewRequest(http.MethodPost, ts.URL+"/echo", bytes.NewBufferString("test"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resp, err := cli.Do(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer resp.Body.Close()

			var buf bytes.Buffer
			_, err = io.Copy(&buf, resp.Body)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got, want := buf.String(), "testtesttest"; got != want {
				t.Errorf("returned %q, want %q", got, want)
			}
		})
	}
}
