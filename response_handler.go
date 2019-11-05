package hx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ResponseHandler func(*http.Client, *http.Request, *http.Response, error) (*http.Response, error)

type ResponseError struct {
	Response *http.Response
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("the server responeded with status %d", e.Response.StatusCode)
}

func AsJSON(dst interface{}) ResponseHandler {
	return func(_ *http.Client, _ *http.Request, r *http.Response, err error) (*http.Response, error) {
		if r == nil {
			return r, err
		}
		defer r.Body.Close()
		err = json.NewDecoder(r.Body).Decode(dst)
		if err != nil {
			return r, err
		}
		return r, nil
	}
}

func AsError() ResponseHandler {
	return func(_ *http.Client, _ *http.Request, r *http.Response, err error) (*http.Response, error) {
		if r == nil || err != nil {
			return r, err
		}
		err = bufferAndCloseResponse(r)
		if err != nil {
			return r, err
		}
		return r, &ResponseError{Response: r}
	}
}

func bufferAndCloseResponse(r *http.Response) error {
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

func checkStatus(f func(int) bool) func(*http.Response, error) bool {
	return func(r *http.Response, err error) bool {
		return err == nil && r != nil && f(r.StatusCode)
	}
}

type ResponseHandlerCond func(*http.Response, error) bool

func IsOK() ResponseHandlerCond { return checkStatus(func(c int) bool { return c/100 == 2 }) }
func IsNotOK() ResponseHandlerCond {
	return func(r *http.Response, err error) bool { return !IsOK()(r, err) }
}
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

func When(cond ResponseHandlerCond, h ResponseHandler) ClientOption {
	return newResponseHandler(func(c *http.Client, req *http.Request, resp *http.Response, err error) (*http.Response, error) {
		if cond(resp, err) {
			return h(c, req, resp, err)
		}
		return resp, err
	})
}

func WhenOK(h ResponseHandler) ClientOption                   { return When(IsOK(), h) }
func WhenNotOK(h ResponseHandler) ClientOption                { return When(IsNotOK(), h) }
func WhenClientError(h ResponseHandler) ClientOption          { return When(IsClientError(), h) }
func WhenServerError(h ResponseHandler) ClientOption          { return When(IsServerError(), h) }
func WhenStatus(h ResponseHandler, codes ...int) ClientOption { return When(IsStatus(codes...), h) }
