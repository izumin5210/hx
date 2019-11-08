package pb

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/izumin5210/hx"
)

func JSON(pb proto.Message) hx.Option {
	return (&JSONConfig{}).JSON(pb)
}

func AsJSON(pb proto.Message) hx.ResponseHandler {
	return (&JSONConfig{}).AsJSON(pb)
}

type JSONConfig struct {
	jsonpb.Marshaler
	jsonpb.Unmarshaler
	EncodeFunc func(proto.Message) (io.Reader, error)
	DecodeFunc func(io.Reader, proto.Message) error
}

func (c *JSONConfig) JSON(pb proto.Message) hx.Option {
	return hx.OptionFunc(func(hc *hx.Config) {
		hc.BodyOption = func(context.Context) (io.Reader, error) {
			return c.encode(pb)
		}
	})
}

func (c *JSONConfig) AsJSON(pb proto.Message) hx.ResponseHandler {
	return func(r *http.Response, err error) (*http.Response, error) {
		if r == nil || err != nil {
			return r, err
		}

		defer r.Body.Close()
		err = c.decode(r.Body, pb)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
}

func (c *JSONConfig) encode(pb proto.Message) (io.Reader, error) {
	if f := c.EncodeFunc; f != nil {
		return f(pb)
	}

	var buf bytes.Buffer
	err := c.Marshal(&buf, pb)
	if err != nil {
		return nil, err
	}
	return &buf, nil
}

func (c *JSONConfig) decode(r io.Reader, pb proto.Message) error {
	if f := c.DecodeFunc; f != nil {
		return f(r, pb)
	}

	return c.Unmarshal(r, pb)
}