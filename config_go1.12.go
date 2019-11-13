// +build !go1.13

package hx

import (
	"context"
	"io"
	"net/http"
)

func init() {
	newRequest = func(ctx context.Context, meth, url string, body io.Reader) (*http.Request, error) {
		req, err := http.NewRequest(meth, url, body)
		if err != nil {
			return nil, err
		}

		return req.WithContext(ctx), nil
	}
}
