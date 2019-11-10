package pb

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/izumin5210/hx"
)

var DefaultProtoConfig = &ProtoConfig{}

// Proto sets proto.Message to request body as protocol buffers.
// This will marshal a given data with proto.Marshal in default.
func Proto(pb proto.Message) hx.Option {
	return DefaultProtoConfig.Proto(pb)
}

// AsProto is hx.ResponseHandler for unmarshaling response bodies as Proto.
// This will unmarshal a received data with proto.Unmarshal marshaler in default.
func AsProto(pb proto.Message) hx.ResponseHandler {
	return DefaultProtoConfig.AsProto(pb)
}

type ProtoConfig struct {
	EncodeFunc func(proto.Message) (io.Reader, error)
	DecodeFunc func(io.Reader, proto.Message) error
}

func (c *ProtoConfig) Proto(pb proto.Message) hx.Option {
	return hx.OptionFunc(func(hc *hx.Config) {
		hc.BodyOption = func(context.Context) (io.Reader, error) {
			return c.encode(pb)
		}
	})
}

func (c *ProtoConfig) AsProto(pb proto.Message) hx.ResponseHandler {
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

func (c *ProtoConfig) encode(pb proto.Message) (io.Reader, error) {
	if f := c.EncodeFunc; f != nil {
		return f(pb)
	}

	data, err := proto.Marshal(pb)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func (c *ProtoConfig) decode(r io.Reader, pb proto.Message) error {
	if f := c.DecodeFunc; f != nil {
		return f(r, pb)
	}

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		return err
	}
	return proto.Unmarshal(buf.Bytes(), pb)
}
