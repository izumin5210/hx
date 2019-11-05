package hx

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

var (
	DefaultUserAgent     = fmt.Sprintf("hx/%s; %s", Version, runtime.Version())
	DefaultClientOptions = []ClientOption{
		UserAgent(DefaultUserAgent),
	}
)

type ClientOption interface {
	Apply(*ClientConfig)
}

type ClientOptionFunc func(*ClientConfig)

func (f ClientOptionFunc) Apply(c *ClientConfig) { f(c) }

func CombineClientOptions(opts ...ClientOption) ClientOption {
	return ClientOptionFunc(func(c *ClientConfig) {
		for _, o := range opts {
			o.Apply(c)
		}
	})
}

func newRequestOption(f func(context.Context, *http.Request) error) ClientOption {
	return ClientOptionFunc(func(c *ClientConfig) {
		c.RequestOptions = append(c.RequestOptions, f)
	})
}

func newURLOption(f func(context.Context, *url.URL) error) ClientOption {
	return ClientOptionFunc(func(c *ClientConfig) {
		c.URLOptions = append(c.URLOptions, f)
	})
}

func newClientOption(f func(context.Context, *http.Client) error) ClientOption {
	return ClientOptionFunc(func(c *ClientConfig) {
		c.ClientOptions = append(c.ClientOptions, f)
	})
}

func newBodyOption(f func(context.Context) (io.Reader, error)) ClientOption {
	return ClientOptionFunc(func(c *ClientConfig) {
		c.BodyOption = f
	})
}

func setBodyOption(r io.Reader) ClientOption {
	return newBodyOption(func(context.Context) (io.Reader, error) { return r, nil })
}

func newResponseHandler(f ResponseHandler) ClientOption {
	return ClientOptionFunc(func(c *ClientConfig) {
		c.ResponseHandlers = append(c.ResponseHandlers, f)
	})
}

func BaseURL(baseURL *url.URL) ClientOption {
	return newURLOption(func(_ context.Context, dest *url.URL) error {
		*dest = *baseURL
		return nil
	})
}

func URL(urlStr string) ClientOption {
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

// BasicAuth sets an username and a password for basic authentication.
func BasicAuth(username, password string) ClientOption {
	return newRequestOption(func(_ context.Context, req *http.Request) error {
		req.SetBasicAuth(username, password)
		return nil
	})
}

// Header sets a value to request header.
func Header(k, v string) ClientOption {
	return newRequestOption(func(_ context.Context, req *http.Request) error {
		req.Header.Set(k, v)
		return nil
	})
}

// Authorization sets an authorization scheme and a token of an user.
func Authorization(scheme, token string) ClientOption {
	return Header("Authorization", scheme+" "+token)
}

// Query sets an url query parameter.
func Query(k, v string) ClientOption {
	return newURLOption(func(_ context.Context, u *url.URL) error {
		q := u.Query()
		q.Set(k, v)
		u.RawQuery = q.Encode()
		return nil
	})
}

// HTTPClient sets a HTTP client that used to send HTTP request(s).
func HTTPClient(cli *http.Client) ClientOption {
	return newClientOption(func(_ context.Context, dest *http.Client) error {
		*dest = *cli
		return nil
	})
}

// Transport sets the round tripper to http.Client.
func Transport(f func(context.Context, http.RoundTripper) http.RoundTripper) ClientOption {
	return newClientOption(func(ctx context.Context, cli *http.Client) error {
		cli.Transport = f(ctx, cli.Transport)
		return nil
	})
}

// Timeout sets the max duration for http request(s).
func Timeout(t time.Duration) ClientOption {
	return newClientOption(func(_ context.Context, cli *http.Client) error {
		cli.Timeout = t
		return nil
	})
}

func UserAgent(ua string) ClientOption {
	return Header("User-Agent", ua)
}

// Body sets data to request body.
func Body(v interface{}) ClientOption {
	switch v := v.(type) {
	case io.Reader:
		return setBodyOption(v)
	case string:
		return setBodyOption(strings.NewReader(v))
	case []byte:
		return setBodyOption(bytes.NewReader(v))
	case url.Values:
		return CombineClientOptions(
			setBodyOption(strings.NewReader(v.Encode())),
			Header("Content-Type", "application/x-www-form-urlencoded"),
		)
	case json.Marshaler:
		return CombineClientOptions(
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

// FormBody sets data to request body as formm.
func FormBody(v interface{}) ClientOption {
	bodyOpt := func() ClientOption {
		switch v := v.(type) {
		case io.Reader:
			return setBodyOption(v)
		case string:
			return setBodyOption(strings.NewReader(v))
		case []byte:
			return setBodyOption(bytes.NewReader(v))
		case url.Values:
			return setBodyOption(strings.NewReader(v.Encode()))
		default:
			return newBodyOption(func(context.Context) (io.Reader, error) {
				return nil, errors.New("failed to encoding request body")
			})
		}
	}()
	return CombineClientOptions(
		bodyOpt,
		Header("Content-Type", "application/x-www-form-urlencoded"),
	)
}

// JSON sets data to request body as json.
func JSON(v interface{}) ClientOption {
	bodyOpt := func() ClientOption {
		switch v := v.(type) {
		case io.Reader:
			return setBodyOption(v)
		case string:
			return setBodyOption(strings.NewReader(v))
		case []byte:
			return setBodyOption(bytes.NewReader(v))
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
	return CombineClientOptions(
		bodyOpt,
		Header("Content-Type", "application/json"),
	)
}
