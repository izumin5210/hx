package hxutil

import (
	"bytes"
	"io/ioutil"
	"net/http"
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
