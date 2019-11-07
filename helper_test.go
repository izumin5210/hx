package hx_test

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/izumin5210/hx"
)

func TestCloneTransport(t *testing.T) {
	// https://github.com/golang/go/blob/go1.13.4/src/net/http/transport.go#L42-L54
	base := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	cloned := hx.CloneTransport(base)
	cloned.MaxIdleConns = 500
	cloned.MaxIdleConnsPerHost = 100

	if cloned.Proxy == nil {
		t.Errorf("Proxy should be copied")
	}

	if cloned.DialContext == nil {
		t.Errorf("DialContext should be copied")
	}

	if got, want := cloned.IdleConnTimeout, base.IdleConnTimeout; got != want {
		t.Errorf("cloned IdleConnTimeout is %s, want %s", got, want)
	}

	if got, want := cloned.TLSHandshakeTimeout, base.TLSHandshakeTimeout; got != want {
		t.Errorf("cloned TLSHandshakeTimeout is %s, want %s", got, want)
	}

	if got, want := cloned.ExpectContinueTimeout, base.ExpectContinueTimeout; got != want {
		t.Errorf("cloned ExpectContinueTimeout is %s, want %s", got, want)
	}

	if got, want := base.MaxIdleConns, 100; got != want {
		t.Errorf("base MaxIdleConns is %d, want %d", got, want)
	}

	if got, want := cloned.MaxIdleConns, 500; got != want {
		t.Errorf("cloned MaxIdleConns is %d, want %d", got, want)
	}

	if got, want := base.MaxIdleConnsPerHost, 0; got != want {
		t.Errorf("base MaxIdleConnsPerHost is %d, want %d", got, want)
	}

	if got, want := cloned.MaxIdleConnsPerHost, 100; got != want {
		t.Errorf("cloned MaxIdleConnsPerHost is %d, want %d", got, want)
	}
}

func TestPath(t *testing.T) {
	cases := []struct {
		test string
		got  string
		want string
	}{
		{
			test: "simple",
			got:  hx.Path("/api/contents", 1, "stargazers"),
			want: "/api/contents/1/stargazers",
		},
		{
			test: "stringer",
			got:  hx.Path("/api", "contents", fakeStringer("fakestringer"), "stargazers"),
			want: "/api/contents/fakestringer/stargazers",
		},
		{
			test: "abs",
			got:  hx.Path("https://api.example.com", "contents", uint32(3), "stargazers"),
			want: "https://api.example.com/contents/3/stargazers",
		},
	}

	for _, tc := range cases {
		t.Run(tc.test, func(t *testing.T) {
			if got, want := tc.got, tc.want; got != want {
				t.Errorf("hx.Path(...) returns %q, want %q", got, want)
			}
		})
	}
}
