package hx

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/izumin5210/hx/hxutil"
)

type ResponseHandler func(*http.Response, error) (*http.Response, error)

func HandleResponse(f func(*http.Response, error) (*http.Response, error)) Option {
	return OptionFunc(func(c *Config) { c.ResponseHandlers = append(c.ResponseHandlers, f) })
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

func AsJSON(dst interface{}) ResponseHandler {
	return func(r *http.Response, err error) (*http.Response, error) {
		if r == nil || err != nil {
			return r, err
		}
		defer r.Body.Close()
		err = json.NewDecoder(r.Body).Decode(dst)
		if err != nil {
			return nil, &ResponseError{Response: r, Err: err}
		}
		return r, nil
	}
}

func AsError() ResponseHandler {
	return func(r *http.Response, err error) (*http.Response, error) {
		if r == nil || err != nil {
			return r, err
		}
		err = hxutil.DrainResponseBody(r)
		if err != nil {
			return nil, &ResponseError{Response: r, Err: err}
		}
		return r, &ResponseError{Response: r}
	}
}

// AsJSONError is ResponseHandler that will populate an error with the JSON returned within the response body.
// And it will wrap the error with ResponseError and return it.
//  err := hx.Post(ctx, "https://example.com/posts",
//  	hx.JSON(body)
//  	hx.WhenSuccess(hx.AsJSON(&post), http.StatusBadRequest),
//  	hx.WhenStatus(hx.AsErrorOf(&InvalidArgument{}), http.StatusBadRequest),
//  	hx.WhenFailure(hx.AsError()),
//  )
//  if err != nil {
//  	var (
//  		invalidArgErr *InvalidArgument
//  		respErr       *hx.ResponseError
//  	)
//  	if errors.As(err, &invalidArgErr) {
//  		// handle known error
//  	} else if errors.As(err, &respErr) {
//  		// handle unknown response error
//  	} else {
//  		err := errors.Unwrap(err)
//  		// handle unknown error
//  	}
//  }
func AsJSONError(dst error) ResponseHandler {
	return func(r *http.Response, err error) (*http.Response, error) {
		if r == nil || err != nil {
			return r, err
		}
		r, err = AsJSON(dst)(r, err)
		if err != nil {
			return nil, &ResponseError{Response: r, Err: err}
		}
		return nil, &ResponseError{Response: r, Err: dst}
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
func IsTemporaryError() ResponseHandlerCond {
	return func(r *http.Response, err error) bool {
		terr, ok := err.(interface{ Temporary() bool })
		return ok && terr.Temporary()
	}
}
func IsStatus(codes ...int) ResponseHandlerCond {
	m := make(map[int]struct{}, len(codes))
	for _, c := range codes {
		m[c] = struct{}{}
	}
	return checkStatus(func(code int) bool { _, ok := m[code]; return ok })
}

func When(cond ResponseHandlerCond, rh ResponseHandler) Option {
	return HandleResponse(func(resp *http.Response, err error) (*http.Response, error) {
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
