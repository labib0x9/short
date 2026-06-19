//go:build e2e
// +build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labib0x9/short/internal/app/url"
	urlhandler "github.com/labib0x9/short/internal/transport/http/handler/url"
)

// Mock service for E2E testing
type mockService struct {
	shortenCalls int
	getCalls     int
}

func (m *mockService) Shorten(ctx interface{}, longUrl string, expireAt *time.Time, userAgent string) (url.ShortenResult, error) {
	m.shortenCalls++
	return url.ShortenResult{
		Msg:      "success",
		Code:     "aB3kZ9",
		ShortUrl: "http://localhost:3000/aB3kZ9",
		ExpireAt: expireAt,
	}, nil
}

func (m *mockService) Get(ctx interface{}, code, referer, userAgent, remoteAddr string) (interface{}, error) {
	m.getCalls++
	return &struct {
		URL string
	}{URL: "https://example.com/path"}, nil
}

func (m *mockService) Analysis(ctx interface{}, code string) (interface{}, error) {
	return nil, nil
}

func (m *mockService) Save(ctx interface{}, msg interface{}) error {
	return nil
}

func (m *mockService) DeleteByExpireAt(ctx interface{}) error {
	return nil
}

func TestE2E_ShortenRequest(t *testing.T) {
	svc := &mockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	// Prepare request
	body := map[string]interface{}{
		"url": "https://www.example.com/very/long/path",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/short", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.Shorten(w, req)

	// Verify
	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["msg"] != "success" {
		t.Errorf("expected msg=success, got %v", resp["msg"])
	}

	if resp["code"] == "" {
		t.Error("expected non-empty code")
	}
}

func TestE2E_InvalidURLRequest(t *testing.T) {
	svc := &mockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	// Invalid URL
	body := map[string]interface{}{
		"url": "not-a-valid-url",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/short", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.Shorten(w, req)

	// Should fail validation
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid URL, got %d", w.Code)
	}
}

func TestE2E_MissingURLRequest(t *testing.T) {
	svc := &mockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	// Missing 'url' field
	body := map[string]interface{}{
		"expire_at": "2025-12-31T23:59:59Z",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/short", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.Shorten(w, req)

	// Should fail validation
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing URL, got %d", w.Code)
	}
}
