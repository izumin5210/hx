package hx

import (
	"fmt"
	"net/http"
)

var (
	_ error = (*RoundTripError)(nil)
	_ error = (*ResponseError)(nil)
)

type RoundTripError struct {
	Err error
}

func (e *RoundTripError) Error() string {
	return fmt.Sprintf("failed to roundtrip: %v", e.Err)
}

func (e *RoundTripError) Unwrap() error {
	return e.Err
}

func isRoundTripError(err error) bool {
	if err == nil {
		return false
	}
	for {
		if _, ok := err.(*RoundTripError); ok {
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

func wrapRoundTripError(r *http.Request, rt http.RoundTripper) (*http.Response, error) {
	resp, err := rt.RoundTrip(r)
	if err != nil {
		err = &RoundTripError{Err: err}
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
