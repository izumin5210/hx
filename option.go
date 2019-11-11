package hx

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/izumin5210/hx/hxutil"
)

var (
	DefaultUserAgent = fmt.Sprintf("hx/%s; %s", Version, runtime.Version())
	DefaultOptions   = []Option{
		UserAgent(DefaultUserAgent),
		TransportFunc(wrapRoundTripError),
	}
)

type Option interface {
	ApplyOption(*Config)
}

type OptionFunc func(*Config)

func (f OptionFunc) ApplyOption(c *Config) { f(c) }

func CombineOptions(opts ...Option) Option {
	return OptionFunc(func(c *Config) {
		for _, o := range opts {
			o.ApplyOption(c)
		}
	})
}

func newURLOption(f func(context.Context, *url.URL) error) Option {
	return OptionFunc(func(c *Config) {
		c.URLOptions = append(c.URLOptions, f)
	})
}

func newBodyOption(f func(context.Context) (io.Reader, error)) Option {
	return OptionFunc(func(c *Config) {
		c.BodyOption = f
	})
}

func setBodyOption(r io.Reader) Option {
	return newBodyOption(func(context.Context) (io.Reader, error) { return r, nil })
}

func newClientOption(f func(context.Context, *http.Client) error) Option {
	return OptionFunc(func(c *Config) {
		c.ClientOptions = append(c.ClientOptions, f)
	})
}

func BaseURL(baseURL *url.URL) Option {
	return newURLOption(func(_ context.Context, dest *url.URL) error {
		*dest = *baseURL
		return nil
	})
}

func URL(urlStr string) Option {
	return newURLOption(func(_ context.Context, base *url.URL) error {
		parse := url.Parse
		if base != nil {
			parse = base.Parse
		}
		newURL, err := parse(urlStr)
		if err != nil {
			return err
		}
		*base = *newURL
		return nil
	})
}

// Query sets an url query parameter.
func Query(k, v string) Option {
	return newURLOption(func(_ context.Context, u *url.URL) error {
		q := u.Query()
		q.Set(k, v)
		u.RawQuery = q.Encode()
		return nil
	})
}

// Body sets data to request body.
func Body(v interface{}) Option {
	switch v := v.(type) {
	case io.Reader:
		return setBodyOption(v)
	case string:
		return setBodyOption(strings.NewReader(v))
	case []byte:
		return setBodyOption(bytes.NewReader(v))
	case url.Values:
		return CombineOptions(
			setBodyOption(strings.NewReader(v.Encode())),
			Header("Content-Type", "application/x-www-form-urlencoded"),
		)
	case json.Marshaler:
		return CombineOptions(
			newBodyOption(func(context.Context) (io.Reader, error) {
				data, err := v.MarshalJSON()
				if err != nil {
					return nil, err
				}
				return bytes.NewReader(data), nil
			}),
			Header("Content-Type", "application/json"),
		)
	case encoding.TextMarshaler:
		return newBodyOption(func(context.Context) (io.Reader, error) {
			data, err := v.MarshalText()
			if err != nil {
				return nil, err
			}
			return bytes.NewReader(data), nil
		})
	case fmt.Stringer:
		return setBodyOption(strings.NewReader(v.String()))
	default:
		return newBodyOption(func(context.Context) (io.Reader, error) {
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(v)
			if err != nil {
				return nil, err
			}
			return &buf, nil
		})
	}
}

// JSON sets data to request body as json.
func JSON(v interface{}) Option {
	bodyOpt := func() Option {
		switch v := v.(type) {
		case io.Reader, string, []byte:
			return Body(v)
		default:
			return newBodyOption(func(context.Context) (io.Reader, error) {
				var buf bytes.Buffer
				err := json.NewEncoder(&buf).Encode(v)
				if err != nil {
					return nil, err
				}
				return &buf, nil
			})
		}
	}()
	return CombineOptions(
		bodyOpt,
		Header("Content-Type", "application/json"),
	)
}

// HTTPClient sets a HTTP client that used to send HTTP request(s).
func HTTPClient(c *http.Client) Option {
	return newClientOption(func(_ context.Context, old *http.Client) error {
		*old = *c
		return nil
	})
}

// Transport sets the round tripper to http.Client.
func Transport(rt http.RoundTripper) Option {
	return CombineOptions(
		newClientOption(func(_ context.Context, c *http.Client) error {
			c.Transport = rt
			return nil
		}),
		TransportFunc(wrapRoundTripError),
	)
}

// TransportFrom sets the round tripper to http.Client.
func TransportFrom(f func(http.RoundTripper) http.RoundTripper) Option {
	return newClientOption(func(_ context.Context, c *http.Client) error {
		c.Transport = f(c.Transport)
		return nil
	})
}

func TransportFunc(f func(*http.Request, http.RoundTripper) (*http.Response, error)) Option {
	return TransportFrom(hxutil.RoundTripperFunc(f).Wrap)
}

// Timeout sets the max duration for http request(s).
func Timeout(t time.Duration) Option {
	return newClientOption(func(_ context.Context, c *http.Client) error {
		c.Timeout = t
		return nil
	})
}
