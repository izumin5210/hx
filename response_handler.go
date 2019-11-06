package hx

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ResponseHandler func(*http.Response, error) (*http.Response, error)

type ResponseError struct {
	Response *http.Response
	err      error
}

func (e *ResponseError) Error() string {
	msg := fmt.Sprintf("the server responeded with status %d", e.Response.StatusCode)
	if e.err != nil {
		msg = fmt.Sprintf("%s: %s", msg, e.err.Error())
	}
	return msg
}

func (e *ResponseError) Unwrap() error {
	if e.err != nil {
		return e.err
	}
	return e
}

func AsJSON(dst interface{}) ResponseHandler {
	return func(r *http.Response, err error) (*http.Response, error) {
		if r == nil || err != nil {
			return r, err
		}
		defer r.Body.Close()
		err = json.NewDecoder(r.Body).Decode(dst)
		if err != nil {
			return nil, &ResponseError{Response: r, err: err}
		}
		return r, nil
	}
}

func AsError() ResponseHandler {
	return func(r *http.Response, err error) (*http.Response, error) {
		if r == nil || err != nil {
			return r, err
		}
		err = DrainResponseBody(r)
		if err != nil {
			return nil, &ResponseError{Response: r, err: err}
		}
		return r, &ResponseError{Response: r}
	}
}

func checkStatus(f func(int) bool) func(*http.Response, error) bool {
	return func(r *http.Response, err error) bool {
		return err == nil && r != nil && f(r.StatusCode)
	}
}

type ResponseHandlerCond func(*http.Response, error) bool

func Any(conds ...ResponseHandlerCond) ResponseHandlerCond {
	return func(r *http.Response, err error) bool {
		for _, c := range conds {
			if c(r, err) {
				return true
			}
		}
		return false
	}
}

func Not(cond ResponseHandlerCond) ResponseHandlerCond {
	return func(r *http.Response, err error) bool { return !cond(r, err) }
}

func IsSuccess() ResponseHandlerCond     { return checkStatus(func(c int) bool { return c/100 == 2 }) }
func IsFailure() ResponseHandlerCond     { return Not(IsSuccess()) }
func IsClientError() ResponseHandlerCond { return checkStatus(func(c int) bool { return c/100 == 4 }) }
func IsServerError() ResponseHandlerCond { return checkStatus(func(c int) bool { return c/100 == 5 }) }
func IsNetworkError() ResponseHandlerCond {
	return func(r *http.Response, err error) bool { return err != nil }
}
func IsStatus(codes ...int) ResponseHandlerCond {
	m := make(map[int]struct{}, len(codes))
	for _, c := range codes {
		m[c] = struct{}{}
	}
	return checkStatus(func(code int) bool { _, ok := m[code]; return ok })
}

func When(cond ResponseHandlerCond, rh ResponseHandler) Option {
	return Then(func(resp *http.Response, err error) (*http.Response, error) {
		if cond(resp, err) {
			return rh(resp, err)
		}
		return resp, err
	})
}

func WhenSuccess(h ResponseHandler) Option              { return When(IsSuccess(), h) }
func WhenFailure(h ResponseHandler) Option              { return When(IsFailure(), h) }
func WhenClientError(h ResponseHandler) Option          { return When(IsClientError(), h) }
func WhenServerError(h ResponseHandler) Option          { return When(IsServerError(), h) }
func WhenStatus(h ResponseHandler, codes ...int) Option { return When(IsStatus(codes...), h) }
