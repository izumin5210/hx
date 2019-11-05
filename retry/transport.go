package retry

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/cenkalti/backoff"
	"github.com/izumin5210/hx"
)

type Transport struct {
	parent http.RoundTripper
	cond   hx.ResponseHandlerCond
	bo     backoff.BackOff
}

var _ http.RoundTripper = (*Transport)(nil)

func NewTransport(
	parent http.RoundTripper,
	cond hx.ResponseHandlerCond,
	bo backoff.BackOff,
) *Transport {
	return &Transport{
		parent: parent,
		cond:   cond,
		bo:     bo,
	}
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	bo := backoff.WithContext(t.bo, req.Context())
	bo.Reset()

	if req.Body != nil {
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			return nil, err
		}
		err = req.Body.Close()
		if err != nil {
			return nil, err
		}
		req.Body = ioutil.NopCloser(&buf)
	}

	_ = backoff.Retry(func() error {
		resp, err = t.parent.RoundTrip(req)
		if t.cond(resp, err) {
			return errors.New("retry")
		}
		return nil
	}, bo)

	return
}
