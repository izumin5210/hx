package logging

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/izumin5210/hx/hxutil"
)

func New() hxutil.RoundTripperFunc {
	return With(log.New(os.Stderr, "[hx] ", log.LstdFlags))
}

func With(l *log.Logger) hxutil.RoundTripperFunc {
	return func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
		t := now()

		l.Printf("Request %s %s %s", req.Proto, req.Method, req.URL.String())

		resp, err := next.RoundTrip(req)

		d := since(t)

		if err != nil {
			l.Printf("Response error: %s: %s %s (%s)", err.Error(), req.Method, req.URL.String(), d.String())
		} else {
			l.Printf("Response %s: %s %s (%s)", resp.Status, req.Method, req.URL.String(), d.String())
		}

		return resp, err
	}
}

var (
	now   = time.Now
	since = time.Since
)
