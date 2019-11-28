package hxzap

import (
	"net/http"
	"time"

	"github.com/izumin5210/hx/hxutil"
	"go.uber.org/zap"
)

func New() hxutil.RoundTripperFunc {
	return With(zap.L().Named("hx"))
}

func With(l *zap.Logger) hxutil.RoundTripperFunc {
	return func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
		t := now()

		l := l.With(
			zap.String("proto", req.Proto),
			zap.String("method", req.Method),
			zap.String("host", req.URL.Host),
			zap.String("path", req.URL.Path),
			zap.Stringer("url", req.URL),
		)

		l.Info("Request", zap.Int64("content_length", req.ContentLength))

		resp, err := next.RoundTrip(req)

		d := since(t)

		if err != nil {
			l.Info("Response error",
				zap.Error(err),
				zap.Duration("response_time", d),
			)
		} else {
			l.Info("Response",
				zap.String("status", resp.Status),
				zap.Int("status_code", resp.StatusCode),
				zap.Int64("content_length", resp.ContentLength),
				zap.Duration("response_time", d),
			)
		}

		return resp, err
	}
}

var (
	now   = time.Now
	since = time.Since
)
