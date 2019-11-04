package httpx

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ClientOption func(*ClientConfig) error

func newRequestOption(f func(context.Context, *http.Request) error) ClientOption {
	return func(c *ClientConfig) error {
		c.RequestOptions = append(c.RequestOptions, f)
		return nil
	}
}

func newURLOption(f func(context.Context, *url.URL) error) ClientOption {
	return func(c *ClientConfig) error {
		c.URLOptions = append(c.URLOptions, f)
		return nil
	}
}

func newClientOption(f func(context.Context, *http.Client) error) ClientOption {
	return func(c *ClientConfig) error {
		c.ClientOptions = append(c.ClientOptions, f)
		return nil
	}
}

func newResponseHandler(f func(*http.Client, *http.Response, error) (*http.Response, error)) ClientOption {
	return func(c *ClientConfig) error {
		c.ResponseHandlers = append(c.ResponseHandlers, f)
		return nil
	}
}

func WithBaseURL(baseURL *url.URL) ClientOption {
	return newURLOption(func(_ context.Context, dest *url.URL) error {
		*dest = *baseURL
		return nil
	})
}

func WithURL(urlStr string) ClientOption {
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

// WithBasicAuth sets an username and a password for basic authentication.
func WithBasicAuth(username, password string) ClientOption {
	return newRequestOption(func(_ context.Context, req *http.Request) error {
		req.SetBasicAuth(username, password)
		return nil
	})
}

// WithHeader sets a value to request header.
func WithHeader(k, v string) ClientOption {
	return newRequestOption(func(_ context.Context, req *http.Request) error {
		req.Header.Set(k, v)
		return nil
	})
}

// WithAuthorization sets an authorization scheme and a token of an user.
func WithAuthorization(scheme, token string) ClientOption {
	return WithHeader("Authorization", scheme+" "+token)
}

// WithQuery sets an url query parameter.
func WithQuery(k, v string) ClientOption {
	return newURLOption(func(_ context.Context, u *url.URL) error {
		q := u.Query()
		q.Set(k, v)
		u.RawQuery = q.Encode()
		return nil
	})
}

// WithHTTPClient sets a HTTP client that used to send HTTP request(s).
func WithHTTPClient(cli *http.Client) ClientOption {
	return newClientOption(func(_ context.Context, dest *http.Client) error {
		*dest = *cli
		return nil
	})
}

// WithTransport sets the round tripper to http.Client.
func WithTransport(f func(context.Context, http.RoundTripper) http.RoundTripper) ClientOption {
	return newClientOption(func(ctx context.Context, cli *http.Client) error {
		cli.Transport = f(ctx, cli.Transport)
		return nil
	})
}

// WithTimeout sets the max duration for http request(s).
func WithTimeout(t time.Duration) ClientOption {
	return newClientOption(func(_ context.Context, cli *http.Client) error {
		cli.Timeout = t
		return nil
	})
}

func WithUserAgent(ua string) ClientOption {
	return WithHeader("User-Agent", ua)
}

// WithBody sets data to request body.
func WithBody(v interface{}) ClientOption {
	return func(c *ClientConfig) error {
		switch v := v.(type) {
		case io.Reader:
			c.Body = v
		case string:
			c.Body = strings.NewReader(v)
		case []byte:
			c.Body = bytes.NewReader(v)
		case url.Values:
			c.Body = strings.NewReader(v.Encode())
			WithHeader("Content-Type", "application/x-www-form-urlencoded")(c)
		case json.Marshaler:
			data, err := v.MarshalJSON()
			if err != nil {
				return err
			}
			c.Body = bytes.NewReader(data)
			WithHeader("Content-Type", "application/json")(c)
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
	}
}

// WithFormBody sets data to request body as formm.
func WithFormBody(v interface{}) ClientOption {
	return func(c *ClientConfig) error {
		switch v := v.(type) {
		case io.Reader:
			c.Body = v
		case string:
			c.Body = strings.NewReader(v)
		case []byte:
			c.Body = bytes.NewReader(v)
		case url.Values:
			c.Body = strings.NewReader(v.Encode())
		default:
			return errors.New("failed to encoding request body")
		}
		WithHeader("Content-Type", "application/x-www-form-urlencoded")(c)
		return nil
	}
}

// WithJSON sets data to request body as json.
func WithJSON(v interface{}) ClientOption {
	return func(c *ClientConfig) error {
		switch v := v.(type) {
		case io.Reader:
			c.Body = v
		case string:
			c.Body = strings.NewReader(v)
		case []byte:
			c.Body = bytes.NewReader(v)
		default:
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(v)
			if err != nil {
				return err
			}
			c.Body = &buf
		}
		WithHeader("Content-Type", "application/json")(c)
		return nil
	}
}

func WithBufferingResponse() ClientOption {
	return newResponseHandler(func(c *http.Client, r *http.Response, err error) (*http.Response, error) {
		if r == nil {
			return r, err
		}
		var buf bytes.Buffer
		_, err = buf.ReadFrom(r.Body)
		if err != nil {
			return r, err
		}
		err = r.Body.Close()
		if err != nil {
			return r, err
		}
		r.Body = ioutil.NopCloser(&buf)
		return r, nil
	})
}

func WithResposneJSON(dst interface{}) ClientOption {
	return newResponseHandler(func(c *http.Client, r *http.Response, err error) (*http.Response, error) {
		if r == nil {
			return r, err
		}
		defer r.Body.Close()
		err = json.NewDecoder(r.Body).Decode(dst)
		if err != nil {
			return r, err
		}
		return r, nil
	})
}
