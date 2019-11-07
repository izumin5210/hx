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
	return hx.OptionFunc(func(c *hx.Config) {
		c.BodyOption = jsonEncoder(pb)
	})
}

func AsJSON(pb proto.Message) hx.ResponseHandler {
	return func(r *http.Response, err error) (*http.Response, error) {
		if r == nil || err != nil {
			return r, err
		}
		defer r.Body.Close()
		err = (&jsonpb.Unmarshaler{}).Unmarshal(r.Body, pb)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
}

func jsonEncoder(pb proto.Message) func(context.Context) (io.Reader, error) {
	return func(ctx context.Context) (io.Reader, error) {
		var buf bytes.Buffer
		err := (&jsonpb.Marshaler{}).Marshal(&buf, pb)
		if err != nil {
			return nil, err
		}
		return &buf, nil
	}
}
