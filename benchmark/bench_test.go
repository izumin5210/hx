package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/dghubble/sling"
	"github.com/go-resty/resty/v2"
	"github.com/izumin5210/hx"
	"github.com/levigross/grequests"
	"github.com/parnurzeal/gorequest"
	"gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman.v2/plugins/body"
)

func setupServer() (string, func()) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/messages":
			w.WriteHeader(http.StatusOK)
			q := r.URL.Query()
			userID, _ := strconv.Atoi(q.Get("user_id"))
			msg := q.Get("message")
			json.NewEncoder(w).Encode(map[string]interface{}{"user_id": userID, "message": msg})
		case r.Method == http.MethodPost && r.URL.Path == "/messages":
			w.WriteHeader(http.StatusCreated)
			io.Copy(w, r.Body)
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&Error{Message: "not found"})
		}
	}))

	return ts.URL + "/messages", ts.Close
}

type Message struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

type Error struct {
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Message }

func BenchmarkPOSTWithJSON(b *testing.B) {
	b.Run("Resty", benchmarkResty_POSTWithJSON)
	b.Run("Sling", benchmarkSling_POSTWithJSON)
	b.Run("Gentleman", benchmarkGentleman_POSTWithJSON)
	b.Run("Gorequest", benchmarkGorequest_POSTWithJSON)
	b.Run("Grequests", benchmarkGrequests_POSTWithJSON)
	b.Run("Hx", benchmarkHx_POSTWithJSON)
	b.Run("NetHTTP", benchmarkNetHTTP_POSTWithJSON)
}

func BenchmarkGETWithQuery(b *testing.B) {
	b.Run("Resty", benchmarkResty_GETWithQuery)
	b.Run("Sling", benchmarkSling_GETWithQuery)
	b.Run("Gentleman", benchmarkGentleman_GETWithQuery)
	b.Run("Gorequest", benchmarkGorequest_GETWithQuery)
	b.Run("Grequests", benchmarkGrequests_GETWithQuery)
	b.Run("Hx", benchmarkHx_GETWithQuery)
	b.Run("NetHTTP", benchmarkNetHTTP_GETWithQuery)
}

func benchmarkResty_GETWithQuery(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		client := resty.New()
		_, err := client.R().
			SetQueryParams(map[string]string{"user_id": fmt.Sprint(i), "message": "It works!"}).
			SetResult(&msg).
			SetError(&Error{}).
			Get(url)
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkResty_POSTWithJSON(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		client := resty.New()
		_, err := client.R().
			SetBody(&Message{UserID: i, Message: "It works!"}).
			SetResult(&msg).
			SetError(&Error{}).
			Post(url)
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkSling_GETWithQuery(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		client := sling.New()
		_, err := client.Get(url).
			QueryStruct(&Message{UserID: i, Message: "It works!"}).
			ReceiveSuccess(&msg) // sling closes a response body automatically
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkSling_POSTWithJSON(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		client := sling.New()
		_, err := client.Post(url).
			BodyJSON(&Message{UserID: i, Message: "It works!"}).
			ReceiveSuccess(&msg) // sling closes a response body automatically
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkGentleman_GETWithQuery(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		client := gentleman.New()
		resp, err := client.Request().
			URL(url).
			SetQueryParams(map[string]string{"user_id": fmt.Sprint(i), "message": "It works!"}).
			Method(http.MethodGet).
			Send()
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
		err = resp.JSON(&msg) // closes a response body
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkGentleman_POSTWithJSON(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		client := gentleman.New()
		resp, err := client.Request().
			URL(url).
			Use(body.JSON(&Message{UserID: i, Message: "It works!"})).
			Method(http.MethodPost).
			Send()
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
		err = resp.JSON(&msg) // closes a response body
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkGorequest_GETWithQuery(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		client := gorequest.New()
		_, _, err := client.Get(url).
			Query(&Message{UserID: i, Message: "It works!"}).
			EndStruct(&msg)
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkGorequest_POSTWithJSON(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		client := gorequest.New()
		_, _, err := client.Post(url).
			Send(&Message{UserID: i, Message: "It works!"}).
			EndStruct(&msg)
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkGrequests_GETWithQuery(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		resp, err := grequests.Get(url, &grequests.RequestOptions{
			QueryStruct: &Message{UserID: i, Message: "It works!"},
		})
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
		err = resp.JSON(&msg) // closes a response body
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkGrequests_POSTWithJSON(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		resp, err := grequests.Post(url, &grequests.RequestOptions{
			JSON: &Message{UserID: i, Message: "It works!"},
		})
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
		err = resp.JSON(&msg) // closes a response body
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkHx_GETWithQuery(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		client := hx.NewClient()
		err := client.Get(context.Background(), url,
			hx.Query("user_id", fmt.Sprint(i)),
			hx.Query("message", "It works!"),
			hx.WhenSuccess(hx.AsJSON(&msg)),
			hx.WhenFailure(hx.AsJSONError(&Error{})),
		)
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkHx_POSTWithJSON(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		client := hx.NewClient()
		err := client.Post(context.Background(), url,
			hx.JSON(&Message{UserID: i, Message: "It works!"}),
			hx.WhenSuccess(hx.AsJSON(&msg)),
			hx.WhenFailure(hx.AsJSONError(&Error{})),
		)
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkNetHTTP_GETWithQuery(b *testing.B) {
	u, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message

		q := url.Values{}
		q.Add("user_id", fmt.Sprint(i))
		q.Add("message", "It works!")

		resp, err := http.Get(u + "?" + q.Encode())
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&msg)
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}

func benchmarkNetHTTP_POSTWithJSON(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		var reqBuf bytes.Buffer
		err := json.NewEncoder(&reqBuf).Encode(&Message{UserID: i, Message: "It works!"})
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
		resp, err := http.Post(url, "application/json", &reqBuf)
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&msg)
		if err != nil {
			b.Errorf("returned %v, want nil", err)
		}
	}
}
