package httpx_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/izumin5210/httpx"
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
	err := httpx.Get(
		ctx,
		ts.URL+"/echo",
		httpx.Query("message", "It Works!"),
		httpx.WhenOK(httpx.AsJSON(&out)),
		httpx.WhenNotOK(httpx.AsError()),
	)
	if err != nil {
		// Handle errors...
	}
	fmt.Println(out.Message)

	// Output:
	// It Works!
}
