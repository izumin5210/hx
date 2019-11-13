package hx

import (
	"net/http"
)

type RequestHandler func(*http.Request) (*http.Request, error)

func HandleRequest(f func(*http.Request) (*http.Request, error)) Option {
	return OptionFunc(func(c *Config) error {
		c.RequestHandlers = append(c.RequestHandlers, f)
		return nil
	})
}

// BasicAuth sets an username and a password for basic authentication.
func BasicAuth(username, password string) Option {
	return HandleRequest(func(r *http.Request) (*http.Request, error) {
		r.SetBasicAuth(username, password)
		return r, nil
	})
}

// Header sets a value to request header.
func Header(k, v string) Option {
	return HandleRequest(func(r *http.Request) (*http.Request, error) {
		r.Header.Set(k, v)
		return r, nil
	})
}

// Authorization sets an authorization scheme and a token of an user.
func Authorization(scheme, token string) Option {
	return Header("Authorization", scheme+" "+token)
}

func Bearer(token string) Option {
	return Authorization("Bearer", token)
}

func UserAgent(ua string) Option {
	return Header("User-Agent", ua)
}
