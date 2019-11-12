package hx

import (
	"bytes"
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
	}

	contentTypeJSON = Header("Content-Type", "application/json")
	contentTypeForm = Header("Content-Type", "application/x-www-form-urlencoded")
)

type Option interface {
	ApplyOption(*Config) error
}

type OptionFunc func(*Config) error

func (f OptionFunc) ApplyOption(c *Config) error { return f(c) }

func CombineOptions(opts ...Option) Option {
	return OptionFunc(func(c *Config) error {
		for _, o := range opts {
			err := o.ApplyOption(c)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func BaseURL(baseURL *url.URL) Option {
	return OptionFunc(func(c *Config) error {
		c.URL = baseURL
		return nil
	})
}

func URL(urlStr string) Option {
	return OptionFunc(func(c *Config) error {
		parse := url.Parse
		if u := c.URL; u != nil {
			parse = u.Parse
		}
		newURL, err := parse(urlStr)
		if err != nil {
			return err
		}
		c.URL = newURL
		return nil
	})
}

// Query sets an url query parameter.
func Query(k, v string) Option {
	return OptionFunc(func(c *Config) error {
		c.QueryParams.Set(k, v)
		return nil
	})
}

// Body sets data to request body.
func Body(v interface{}) Option {
	return OptionFunc(func(c *Config) error {
		switch v := v.(type) {
		case io.Reader:
			c.Body = v
		case string:
			c.Body = strings.NewReader(v)
		case []byte:
			c.Body = bytes.NewReader(v)
		case url.Values:
			c.Body = strings.NewReader(v.Encode())
			err := contentTypeForm.ApplyOption(c)
			if err != nil {
				return err
			}
		case json.Marshaler:
			data, err := v.MarshalJSON()
			if err != nil {
				return err
			}
			c.Body = bytes.NewReader(data)
			err = contentTypeJSON.ApplyOption(c)
			if err != nil {
				return err
			}
		case encoding.TextMarshaler:
			data, err := v.MarshalText()
			if err != nil {
				return err
			}
			c.Body = bytes.NewReader(data)
		case fmt.Stringer:
			c.Body = strings.NewReader(v.String())
		default:
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(v)
			if err != nil {
				return err
			}
			c.Body = &buf
		}
		return nil
	})
}

// JSON sets data to request body as json.
func JSON(v interface{}) Option {
	return OptionFunc(func(c *Config) error {
		switch v := v.(type) {
		case io.Reader, string, []byte:
			err := Body(v).ApplyOption(c)
			if err != nil {
				return err
			}
		default:
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(v)
			if err != nil {
				return err
			}
			c.Body = &buf
		}
		return contentTypeJSON.ApplyOption(c)
	})
}

// HTTPClient sets a HTTP client that used to send HTTP request(s).
func HTTPClient(cli *http.Client) Option {
	return OptionFunc(func(c *Config) error {
		c.HTTPClient = cli
		return nil
	})
}

// Transport sets the round tripper to http.Client.
func Transport(rt http.RoundTripper) Option {
	return OptionFunc(func(c *Config) error {
		c.HTTPClient.Transport = rt
		return nil
	})
}

// TransportFrom sets the round tripper to http.Client.
func TransportFrom(f func(http.RoundTripper) http.RoundTripper) Option {
	return OptionFunc(func(c *Config) error {
		c.HTTPClient.Transport = f(c.HTTPClient.Transport)
		return nil
	})
}

func TransportFunc(f func(*http.Request, http.RoundTripper) (*http.Response, error)) Option {
	return TransportFrom(hxutil.RoundTripperFunc(f).Wrap)
}

// Timeout sets the max duration for http request(s).
func Timeout(t time.Duration) Option {
	return OptionFunc(func(c *Config) error {
		c.HTTPClient.Timeout = t
		return nil
	})
}
