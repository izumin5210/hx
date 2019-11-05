package hx

import (
	"net/http"
)

type Handler func(*http.Client, *http.Request) (*http.Response, error)
type Interceptor func(*http.Client, *http.Request, Handler) (*http.Response, error)

func combineInterceptors(interceptors []Interceptor) Interceptor {
	n := len(interceptors)

	return func(cli *http.Client, req *http.Request, h Handler) (*http.Response, error) {
		next := h

		for i := n - 1; i >= 0; i-- {
			next = wrapHandler(interceptors[i], next)
		}

		return next(cli, req)
	}
}

func wrapHandler(i Interceptor, h Handler) Handler {
	return func(c *http.Client, r *http.Request) (*http.Response, error) {
		return i(c, r, h)
	}
}
