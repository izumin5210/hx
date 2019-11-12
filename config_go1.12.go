// +build !go1.13

package hx

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func newRequest(ctx context.Context, meth string, url *url.URL, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(meth, url.String(), body)
	if err != nil {
		return nil, err
	}

	return req.WithContext(ctx), nil
}
