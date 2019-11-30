package hx

import "net/http"

func Intercept(i Interceptor) Option {
	return OptionFunc(func(c *Config) error { c.Interceptors = append(c.Interceptors, i); return nil })
}

func InterceptFunc(f func(*http.Client, *http.Request, RequestFunc) (*http.Response, error)) Option {
	return Intercept(InterceptorFunc(f))
}

type RequestFunc = func(*http.Client, *http.Request) (*http.Response, error)

type Interceptor interface {
	DoRequest(*http.Client, *http.Request, RequestFunc) (*http.Response, error)
	Wrap(RequestFunc) RequestFunc
}

var _ Interceptor = InterceptorFunc(nil)

type InterceptorFunc func(*http.Client, *http.Request, RequestFunc) (*http.Response, error)

func (i InterceptorFunc) DoRequest(c *http.Client, r *http.Request, next RequestFunc) (*http.Response, error) {
	return i(c, r, next)
}

func (i InterceptorFunc) Wrap(f RequestFunc) RequestFunc {
	return func(c *http.Client, r *http.Request) (*http.Response, error) { return i.DoRequest(c, r, f) }
}

func combineInterceptors(interceptors []Interceptor) Interceptor {
	n := len(interceptors)

	return InterceptorFunc(func(cli *http.Client, req *http.Request, f RequestFunc) (*http.Response, error) {
		next := f

		for i := n - 1; i >= 0; i-- {
			next = interceptors[i].Wrap(next)
		}

		return next(cli, req)
	})
}
