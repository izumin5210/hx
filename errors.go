package hx

import (
	"fmt"
	"net/http"
)

var (
	_ error = (*NetworkError)(nil)
	_ error = (*ResponseError)(nil)
)

type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("failed to roundtrip: %v", e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	for {
		if _, ok := err.(*NetworkError); ok {
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return false
}

func wrapNetworkError(r *http.Request, rt http.RoundTripper) (*http.Response, error) {
	resp, err := rt.RoundTrip(r)
	if err != nil {
		err = &NetworkError{Err: err}
	}

	return resp, err
}

type ResponseError struct {
	Response *http.Response
	Err      error
}

func (e *ResponseError) Error() string {
	msg := fmt.Sprintf("the server responeded with status %d", e.Response.StatusCode)
	if e.Err != nil {
		msg = fmt.Sprintf("%s: %s", msg, e.Err.Error())
	}
	return msg
}

func (e *ResponseError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return e
}
