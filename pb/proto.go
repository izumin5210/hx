package pb

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/izumin5210/hx"
)

func Proto(pb proto.Message) hx.Option {
	return hx.OptionFunc(func(c *hx.Config) {
		c.BodyOption = protoEncoder(pb)
	})
}

func AsProto(pb proto.Message) hx.ResponseHandler {
	return func(r *http.Response, err error) (*http.Response, error) {
		if r == nil || err != nil {
			return r, err
		}
		defer r.Body.Close()
		var buf bytes.Buffer
		_, err = io.Copy(&buf, r.Body)
		if err != nil {
			return nil, err
		}
		err = proto.Unmarshal(buf.Bytes(), pb)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
}

func protoEncoder(pb proto.Message) func(context.Context) (io.Reader, error) {
	return func(ctx context.Context) (io.Reader, error) {
		data, err := proto.Marshal(pb)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(data), nil
	}
}
