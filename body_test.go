package hx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/izumin5210/hx"
)

type fakeStringer string

func (s fakeStringer) String() string { return string(s) }

type fakeTextMarshaler string

func (tm fakeTextMarshaler) MarshalText() ([]byte, error) { return []byte(tm), nil }

type fakeJSONMarshaler string

func (jm fakeJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"message":"%s"}`, jm)), nil
}

func TestBody(t *testing.T) {
	type Post struct {
		Message string `json:"message"`
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/echo":
			var msg string
			switch r.Header.Get("Content-Type") {
			case "application/json":
				var post Post
				err := json.NewDecoder(r.Body).Decode(&post)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				msg = post.Message
			case "application/x-www-form-urlencoded":
				err := r.ParseForm()
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				msg = r.PostFormValue("message")
			default:
				data, err := ioutil.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				msg = string(data)
			}
			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(map[string]string{"message": msg})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cases := []struct {
		test string
		in   interface{}
	}{
		{
			test: "with io.Reader",
			in:   bytes.NewBufferString("Hello!"),
		},
		{
			test: "with string",
			in:   "Hello!",
		},
		{
			test: "with bytes",
			in:   []byte("Hello!"),
		},
		{
			test: "with url.Values",
			in:   url.Values{"message": []string{"Hello!"}},
		},
		{
			test: "with fmt.Stringer",
			in:   fakeStringer("Hello!"),
		},
		{
			test: "with encoding.TextMarshaler",
			in:   fakeTextMarshaler("Hello!"),
		},
		{
			test: "with json.Marshaler",
			in:   fakeJSONMarshaler("Hello!"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.test, func(t *testing.T) {
			var out Post
			err := hx.Post(context.Background(), ts.URL+"/echo",
				hx.Body(tc.in),
				hx.WhenSuccess(hx.AsJSON(&out)),
				hx.WhenFailure(hx.AsError()),
			)
			if err != nil {
				t.Errorf("returned %v, want nil", err)
			}
			if got, want := out.Message, "Hello!"; got != want {
				t.Errorf("returned message is %q, want %q", got, want)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/echo":
			out := make(map[string]interface{})
			err := json.NewDecoder(r.Body).Decode(&out)
			if err != nil || out["message"] == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			err = json.NewEncoder(w).Encode(out)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	type Post struct {
		Message string `json:"message"`
	}

	cases := []struct {
		test string
		in   interface{}
	}{
		{
			test: "with struct",
			in:   &Post{Message: "Hello!"},
		},
		{
			test: "with string",
			in:   `{"message": "Hello!"}`,
		},
		{
			test: "with bytes",
			in:   []byte(`{"message": "Hello!"}`),
		},
		{
			test: "with io.Reader",
			in:   bytes.NewBufferString(`{"message": "Hello!"}`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.test, func(t *testing.T) {
			var out Post
			err := hx.Post(context.Background(), ts.URL+"/echo",
				hx.JSON(tc.in),
				hx.WhenSuccess(hx.AsJSON(&out)),
				hx.WhenFailure(hx.AsError()),
			)
			if err != nil {
				t.Errorf("returned %v, want nil", err)
			}
			if got, want := out.Message, "Hello!"; got != want {
				t.Errorf("returned message is %q, want %q", got, want)
			}
		})
	}

	t.Run("failed to encode request body", func(t *testing.T) {
		var out Post
		err := hx.Post(context.Background(), ts.URL+"/echo",
			hx.JSON(func() {}),
			hx.WhenSuccess(hx.AsJSON(&out)),
			hx.WhenFailure(hx.AsError()),
		)
		if err == nil {
			t.Error("returned nil, want an error")
		}
		if _, ok := err.(*hx.ResponseError); ok {
			t.Errorf("returned RequestError %v, want json error", err)
		}
	})
}
