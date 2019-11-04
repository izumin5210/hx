package hx

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type ResponseHandler func(*http.Client, *http.Response, error) (*http.Response, error)

type ResponseError struct {
	Response *http.Response
}

func (e *ResponseError) Error() string {
	return "" // TODO
}

func AsJSON(dst interface{}) ResponseHandler {
	return func(c *http.Client, r *http.Response, err error) (*http.Response, error) {
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
	return func(c *http.Client, r *http.Response, err error) (*http.Response, error) {
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

type respHandlerCond func(*http.Response, error) bool

func newRespHandlerWithCond(f ResponseHandler, cond respHandlerCond) ClientOption {
	return newResponseHandler(func(c *http.Client, r *http.Response, err error) (*http.Response, error) {
		if cond(r, err) {
			return f(c, r, err)
		}
		return r, err
	})
}

func checkStatus(f func(int) bool) func(*http.Response, error) bool {
	return func(r *http.Response, err error) bool {
		return err == nil && r != nil && f(r.StatusCode)
	}
}

var (
	isOK          = checkStatus(func(c int) bool { return c/100 == 2 })
	isNotOK       = func(r *http.Response, err error) bool { return !isOK(r, err) }
	isClientError = checkStatus(func(c int) bool { return c/100 == 4 })
	isServerError = checkStatus(func(c int) bool { return c/100 == 5 })
)

func WhenOK(h ResponseHandler) ClientOption {
	return newRespHandlerWithCond(h, isOK)
}

func WhenNotOK(h ResponseHandler) ClientOption {
	return newRespHandlerWithCond(h, isNotOK)
}

func WhenClientError(h ResponseHandler) ClientOption {
	return newRespHandlerWithCond(h, isClientError)
}

func WhenServerError(h ResponseHandler) ClientOption {
	return newRespHandlerWithCond(h, isServerError)
}

func WhenStatus(h ResponseHandler, codes ...int) ClientOption {
	m := make(map[int]struct{}, len(codes))
	for _, c := range codes {
		m[c] = struct{}{}
	}
	isStatus := checkStatus(func(code int) bool { _, ok := m[code]; return ok })

	return newRespHandlerWithCond(h, isStatus)
}
