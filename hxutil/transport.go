package hxutil

import (
	"net/http"
	"reflect"
)

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
