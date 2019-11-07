package hx

import (
	"net/http"
	"time"
)

type RequestHandler func(*http.Client, *http.Request) (*http.Client, *http.Request, error)

func (h RequestHandler) HandleRequest(c *http.Client, r *http.Request) (*http.Client, *http.Request, error) {
	return h(c, r)
}

func (h RequestHandler) Apply(c *Config) {
	c.RequestHandlers = append(c.RequestHandlers, h)
}

// HTTPClient sets a HTTP client that used to send HTTP request(s).
func HTTPClient(c *http.Client) Option {
	return RequestHandler(func(_ *http.Client, r *http.Request) (*http.Client, *http.Request, error) {
		return c, r, nil
	})
}

// Transport sets the round tripper to http.Client.
func Transport(rt http.RoundTripper) Option {
	return RequestHandler(func(c *http.Client, r *http.Request) (*http.Client, *http.Request, error) {
		c.Transport = rt
		return c, r, nil
	})
}

// TransportFrom sets the round tripper to http.Client.
func TransportFrom(f func(http.RoundTripper) http.RoundTripper) Option {
	return RequestHandler(func(c *http.Client, r *http.Request) (*http.Client, *http.Request, error) {
		c.Transport = f(c.Transport)
		return c, r, nil
	})
}

// Timeout sets the max duration for http request(s).
func Timeout(t time.Duration) Option {
	return RequestHandler(func(c *http.Client, r *http.Request) (*http.Client, *http.Request, error) {
		c.Timeout = t
		return c, r, nil
	})
}

// BasicAuth sets an username and a password for basic authentication.
func BasicAuth(username, password string) Option {
	return RequestHandler(func(c *http.Client, r *http.Request) (*http.Client, *http.Request, error) {
		r.SetBasicAuth(username, password)
		return c, r, nil
	})
}

// Header sets a value to request header.
func Header(k, v string) Option {
	return RequestHandler(func(c *http.Client, r *http.Request) (*http.Client, *http.Request, error) {
		r.Header.Set(k, v)
		return c, r, nil
	})
}
