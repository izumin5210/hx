package pb_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/proto/proto3_proto"
	"github.com/google/go-cmp/cmp"
	"github.com/izumin5210/hx"
	"github.com/izumin5210/hx/plugins/pb"
)

func TestProto(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/echo":
			var (
				msg proto3_proto.Message
				buf bytes.Buffer
			)

			_, err := io.Copy(&buf, r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err = proto.Unmarshal(buf.Bytes(), &msg)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			data, err := proto.Marshal(&msg)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Write(data)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	want := &proto3_proto.Message{
		Name:     "It, Works!",
		Score:    120,
		Hilarity: proto3_proto.Message_SLAPSTICK,
		Children: []*proto3_proto.Message{
			{Name: "foo", HeightInCm: 170},
			{Name: "bar", TrueScotsman: true},
		},
	}

	t.Run("simple", func(t *testing.T) {
		var got proto3_proto.Message
		err := hx.Post(context.Background(), ts.URL+"/echo",
			pb.Proto(want),
			hx.WhenSuccess(pb.AsProto(&got)),
			hx.WhenFailure(hx.AsError()),
		)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
		assertProtoMessage(t, want, &got)
	})

	t.Run("custom encoder", func(t *testing.T) {
		var got, overwrited proto3_proto.Message
		overwrited = *want
		overwrited.Name = "It, Works!!!!!!!!!!!!!!!!!!!!!!"

		protoCfg := &pb.ProtoConfig{
			EncodeFunc: func(_ proto.Message) (io.Reader, error) {
				data, err := proto.Marshal(&overwrited)
				if err != nil {
					return nil, err
				}
				return bytes.NewReader(data), nil
			},
		}
		err := hx.Post(context.Background(), ts.URL+"/echo",
			protoCfg.Proto(want),
			hx.WhenSuccess(pb.AsProto(&got)),
			hx.WhenFailure(hx.AsError()),
		)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
		assertProtoMessage(t, &overwrited, &got)
	})

	t.Run("custom decoder", func(t *testing.T) {
		var got, overwrited proto3_proto.Message
		overwrited = *want
		overwrited.Name = "It, Works!!!!!!!!!!!!!!!!!!!!!!"

		protoCfg := &pb.ProtoConfig{
			DecodeFunc: func(r io.Reader, m proto.Message) error {
				(*m.(*proto3_proto.Message)) = *want
				return nil
			},
		}
		err := hx.Post(context.Background(), ts.URL+"/echo",
			pb.Proto(&overwrited),
			hx.WhenSuccess(protoCfg.AsProto(&got)),
			hx.WhenFailure(hx.AsError()),
		)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
		assertProtoMessage(t, want, &got)
	})
}

func assertProtoMessage(t *testing.T, want proto.Message, got proto.Message) {
	t.Helper()
	if diff := cmp.Diff(want, got, cmp.FilterPath(
		func(p cmp.Path) bool { return !strings.HasPrefix(p.Last().String(), "XXX_") },
		cmp.Ignore(),
	)); diff != "" {
		t.Errorf("response mismatch(-want +got)\n%s", diff)
	}
}
