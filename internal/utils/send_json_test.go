package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendJson(t *testing.T) {
	t.Run("writes status code and JSON body", func(t *testing.T) {
		type payload struct {
			Msg  string `json:"msg"`
			Code int    `json:"code"`
		}

		w := httptest.NewRecorder()
		SendJson(w, payload{Msg: "ok", Code: 42}, http.StatusCreated)

		res := w.Result()

		if res.StatusCode != http.StatusCreated {
			t.Errorf("want status 201, got %d", res.StatusCode)
		}

		if ct := res.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("want Content-Type application/json, got %q", ct)
		}

		var got payload
		if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}

		if got.Msg != "ok" || got.Code != 42 {
			t.Errorf("unexpected body: %+v", got)
		}
	})

	t.Run("handles un-encodable value gracefully", func(t *testing.T) {
		w := httptest.NewRecorder()
		// channels cannot be JSON-encoded
		SendJson(w, make(chan int), http.StatusOK)

		// WriteHeader was already called before the encode error,
		// so the status stays 200 — but the body should be the error message.
		if w.Body.Len() == 0 {
			t.Error("expected non-empty body on encode error")
		}
	})
}

func TestSendError(t *testing.T) {
	t.Run("wraps message in error envelope", func(t *testing.T) {
		w := httptest.NewRecorder()
		SendError(w, "not found", http.StatusNotFound)

		res := w.Result()

		if res.StatusCode != http.StatusNotFound {
			t.Errorf("want 404, got %d", res.StatusCode)
		}

		var got map[string]any
		if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
			t.Fatalf("decode failed: %v", err)
		}

		if got["error"] != "not found" {
			t.Errorf("unexpected error field: %v", got["error"])
		}

		// JSON numbers decode as float64 by default
		if got["code"] != float64(http.StatusNotFound) {
			t.Errorf("unexpected code field: %v", got["code"])
		}
	})
}

type unserializable struct{}

func (u unserializable) MarshalJSON() ([]byte, error) {
	return nil, errors.New("intentional marshal failure")
}

func TestSendJson_EncodeFails(t *testing.T) {
	w := httptest.NewRecorder()
	SendJson(w, unserializable{}, http.StatusOK)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", w.Code)
	}
}
