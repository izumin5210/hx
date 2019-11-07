package hx

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
)

func DrainResponseBody(r *http.Response) error {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		return err
	}
	err = r.Body.Close()
	if err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

// CloneTransport creates a new *http.Transport object that has copied attributes from a given one.
func CloneTransport(in *http.Transport) *http.Transport {
	out := new(http.Transport)
	outRv := reflect.ValueOf(out).Elem()

	rv := reflect.ValueOf(in).Elem()
	rt := rv.Type()

	n := rt.NumField()
	for i := 0; i < n; i++ {
		src, dst := rv.Field(i), outRv.Field(i)
		if src.Type().AssignableTo(dst.Type()) && dst.CanSet() {
			dst.Set(src)
		}
	}

	return out
}

func Path(elem ...interface{}) string {
	chunks := make([]string, len(elem))
	for i, e := range elem {
		var s string
		switch v := e.(type) {
		case string:
			s = v
		case fmt.Stringer:
			s = v.String()
		default:
			s = fmt.Sprint(v)
		}
		chunks[i] = s
	}
	if u, err := url.Parse(chunks[0]); err == nil && u.IsAbs() {
		return strings.TrimSuffix(chunks[0], "/") + "/" + path.Join(chunks[1:]...)
	}
	return path.Join(chunks...)
}
