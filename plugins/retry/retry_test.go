package retry_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/izumin5210/hx"
	"github.com/izumin5210/hx/plugins/retry"
)

func TestRetry(t *testing.T) {
	type Message struct {
		UserID int    `json:"user_id"`
		Body   string `json:"body"`
	}
	var (
		failCount   = 2
		gotMessages []Message
	)
	idempotencyKeys := map[string]interface{}{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/messages":
			if failCount > 0 {
				failCount--
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			failCount--
			if _, ok := idempotencyKeys[r.Header.Get("Idempotency-Key")]; ok {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			var msg Message
			json.NewDecoder(r.Body).Decode(&msg)
			gotMessages = append(gotMessages, msg)

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(&msg)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 50 * time.Millisecond
	bo.MaxInterval = 500 * time.Millisecond

	in := Message{
		UserID: 123,
		Body:   "Hello!",
	}
	var out Message

	err := hx.Post(context.Background(), ts.URL+"/messages",
		retry.When(hx.Any(hx.IsServerError, hx.IsTemporaryError), bo),
		hx.JSON(&in),
		hx.WhenSuccess(hx.AsJSON(&out)),
		hx.WhenFailure(hx.AsError()),
	)
	if err != nil {
		t.Errorf("returned %v, want nil", err)
	}

	if got, want := in, out; !reflect.DeepEqual(got, want) {
		t.Errorf("returned %v, want %v", got, want)
	}

	if got, want := gotMessages, []Message{in}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
