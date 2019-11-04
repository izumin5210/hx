package hx_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/izumin5210/hx"
)

func ExampleGet() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/echo":
			err := json.NewEncoder(w).Encode(map[string]string{
				"message": r.URL.Query().Get("message"),
			})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	var out struct {
		Message string `json:"message"`
	}

	ctx := context.Background()
	err := hx.Get(
		ctx,
		ts.URL+"/echo",
		hx.Query("message", "It Works!"),
		hx.WhenOK(httpx.AsJSON(&out)),
		hx.WhenNotOK(httpx.AsError()),
	)
	if err != nil {
		// Handle errors...
	}
	fmt.Println(out.Message)

	// Output:
	// It Works!
}
