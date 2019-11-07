package hx

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"runtime"
	"strings"
)

var (
	DefaultUserAgent = fmt.Sprintf("hx/%s; %s", Version, runtime.Version())
	DefaultOptions   = []Option{
		UserAgent(DefaultUserAgent),
	}
)

type Option interface {
	Apply(*Config)
}

type OptionFunc func(*Config)

func (f OptionFunc) Apply(c *Config) { f(c) }

func CombineOptions(opts ...Option) Option {
	return OptionFunc(func(c *Config) {
		for _, o := range opts {
			o.Apply(c)
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

// Authorization sets an authorization scheme and a token of an user.
func Authorization(scheme, token string) Option {
	return Header("Authorization", scheme+" "+token)
}

func Bearer(token string) Option {
	return Authorization("Bearer", token)
}

func UserAgent(ua string) Option {
	return Header("User-Agent", ua)
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

// FormBody sets data to request body as formm.
func FormBody(v interface{}) Option {
	bodyOpt := func() Option {
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
	return CombineOptions(
		bodyOpt,
		Header("Content-Type", "application/x-www-form-urlencoded"),
	)
}

// JSON sets data to request body as json.
func JSON(v interface{}) Option {
	bodyOpt := func() Option {
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
	return CombineOptions(
		bodyOpt,
		Header("Content-Type", "application/json"),
	)
}
