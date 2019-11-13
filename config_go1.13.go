// +build go1.13

package hx

import (
	"net/http"
)

func init() {
	newRequest = http.NewRequestWithContext
}
