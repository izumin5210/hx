package hx

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

var DefaultJSONConfig = &JSONConfig{}

type JSONConfig struct {
	EncodeFunc func(interface{}) (io.Reader, error)
	DecodeFunc func(io.Reader, interface{}) error
}

func (c *JSONConfig) JSON(v interface{}) Option {
	return OptionFunc(func(cfg *Config) error {
		r, err := c.encode(v)
		if err != nil {
			return err
		}
		cfg.Body = r
		return contentTypeJSON.ApplyOption(cfg)
	})
}

func (c *JSONConfig) AsJSON(v interface{}) ResponseHandler {
	return func(r *http.Response, err error) (*http.Response, error) {
		if r == nil || err != nil {
			return r, err
		}

		defer r.Body.Close()
		err = c.decode(r.Body, v)
		if err != nil {
			return nil, &ResponseError{Response: r, Err: err}
		}
		return r, nil
	}
}

func (c *JSONConfig) AsJSONError(dst error) ResponseHandler {
	return func(r *http.Response, err error) (*http.Response, error) {
		if r == nil || err != nil {
			return r, err
		}
		err = c.decode(r.Body, dst)
		if err != nil {
			return nil, &ResponseError{Response: r, Err: err}
		}
		return nil, &ResponseError{Response: r, Err: dst}
	}
}

func (c *JSONConfig) encode(v interface{}) (io.Reader, error) {
	if f := c.EncodeFunc; f != nil {
		return f(v)
	}

	switch v := v.(type) {
	case io.Reader:
		return v, nil
	case string:
		return strings.NewReader(v), nil
	case []byte:
		return bytes.NewReader(v), nil
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(data), nil
	}
}

func (c *JSONConfig) decode(r io.Reader, v interface{}) error {
	if f := c.DecodeFunc; f != nil {
		return f(r, v)
	}

	return json.NewDecoder(r).Decode(v)
}
