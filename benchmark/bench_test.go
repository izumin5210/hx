package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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

func BenchmarkResty(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()

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

func BenchmarkSling(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()

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

func BenchmarkGentleman(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()

	for i := 0; i < b.N; i++ {
		var msg Message
		client := gentleman.New()
		resp, err := client.Request().
			URL(url).
			Use(body.JSON(&Message{UserID: i, Message: "It works!"})).
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

func BenchmarkGorequest(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()

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

func BenchmarkGrequests(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()

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

func BenchmarkHx(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()

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

func BenchmarkNetHTTP(b *testing.B) {
	url, closeServer := setupServer()
	defer closeServer()

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
