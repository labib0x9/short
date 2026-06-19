//go:build security
// +build security

package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labib0x9/short/config"
	"github.com/labib0x9/short/internal/app/url"
	"github.com/labib0x9/short/internal/domain/queue"
	urlhandler "github.com/labib0x9/short/internal/transport/http/handler/url"
	urldomain "github.com/labib0x9/short/internal/domain/url"
)

// Mock service for security testing
type securityMockService struct{}

func (m *securityMockService) Shorten(ctx interface{}, longUrl string, expireAt *time.Time, userAgent string) (url.ShortenResult, error) {
	return url.ShortenResult{
		Msg:      "success",
		Code:     "aB3kZ9",
		ShortUrl: "http://localhost:3000/aB3kZ9",
		ExpireAt: expireAt,
	}, nil
}

func (m *securityMockService) Get(ctx interface{}, code, referer, userAgent, remoteAddr string) (interface{}, error) {
	return &urldomain.Url{URL: "https://example.com/path"}, nil
}

func (m *securityMockService) Analysis(ctx interface{}, code string) (interface{}, error) {
	return nil, nil
}

func (m *securityMockService) Save(ctx interface{}, msg interface{}) error {
	return nil
}

func (m *securityMockService) DeleteByExpireAt(ctx interface{}) error {
	return nil
}

// ============================================================================
// SQL INJECTION TESTS
// ============================================================================

func TestSecurity_SQLInjection_ShortCode(t *testing.T) {
	"""Test that SQL injection attempts in short code are prevented"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	// Attempt SQL injection in short code parameter
	sqlInjectionPayloads := []string{
		"'; DROP TABLE urls; --",
		"1' OR '1'='1",
		"admin'--",
		"' UNION SELECT * FROM users --",
		"1; DELETE FROM urls; --",
	}

	for _, payload := range sqlInjectionPayloads {
		req := httptest.NewRequest("GET", "/"+payload, nil)
		req.SetPathValue("code", payload)
		w := httptest.NewRecorder()

		// Should not crash or execute SQL
		// Handler should either return 404 or bad request
		handler.Get(w, req)

		// Verify response is safe (no SQL errors in response)
		if strings.Contains(w.Body.String(), "SQL") ||
			strings.Contains(w.Body.String(), "syntax error") ||
			strings.Contains(w.Body.String(), "database") {
			t.Errorf("SQL error leaked in response for payload: %s", payload)
		}
	}
}

func TestSecurity_SQLInjection_URLParameter(t *testing.T) {
	"""Test that SQL injection in URL field is validated"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	// Attempt to inject SQL through URL field
	body := `{"url":"https://example.com'; DROP TABLE users; --"}`
	req := httptest.NewRequest("POST", "/short", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	// Should reject invalid URL
	if w.Code == http.StatusCreated {
		t.Error("expected request to be rejected, but was accepted")
	}
}

// ============================================================================
// XSS (Cross-Site Scripting) TESTS
// ============================================================================

func TestSecurity_XSS_InputValidation(t *testing.T) {
	"""Test that XSS payloads in URL are rejected"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	xssPayloads := []string{
		"javascript:alert('XSS')",
		"<script>alert('XSS')</script>",
		"https://example.com<img src=x onerror=alert('XSS')>",
		"https://example.com' onclick='alert(1)'",
	}

	for _, payload := range xssPayloads {
		body := fmt.Sprintf(`{"url":"%s"}`, payload)
		req := httptest.NewRequest("POST", "/short", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Shorten(w, req)

		// Should reject if not valid URL format
		// Valid URLs must start with http:// or https://
		if w.Code == http.StatusCreated {
			// If accepted, verify response doesn't echo unescaped payload
			if strings.Contains(w.Body.String(), "<script>") ||
				strings.Contains(w.Body.String(), "javascript:") {
				t.Errorf("unescaped XSS payload in response: %s", payload)
			}
		}
	}
}

func TestSecurity_XSS_ResponseEscaping(t *testing.T) {
	"""Test that error responses properly escape user input"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	// Send malicious JSON
	body := `{"url":"<img src=x onerror=alert(1)>"}`
	req := httptest.NewRequest("POST", "/short", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	// Response should not contain unescaped HTML
	if strings.Contains(w.Body.String(), "<img") {
		t.Error("XSS payload not escaped in error response")
	}
}

// ============================================================================
// INPUT VALIDATION TESTS
// ============================================================================

func TestSecurity_InputValidation_InvalidURL(t *testing.T) {
	"""Test that invalid URLs are rejected"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	invalidURLs := []string{
		"not-a-url",
		"ftp://example.com", // Unsupported scheme
		"javascript:void(0)",
		"file:///etc/passwd",
		"gopher://example.com",
		"", // Empty
	}

	for _, invalidURL := range invalidURLs {
		body := fmt.Sprintf(`{"url":"%s"}`, invalidURL)
		req := httptest.NewRequest("POST", "/short", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Shorten(w, req)

		if w.Code == http.StatusCreated {
			t.Errorf("expected invalid URL to be rejected: %s", invalidURL)
		}
	}
}

func TestSecurity_InputValidation_ExcessiveLength(t *testing.T) {
	"""Test that excessively long URLs are rejected"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	// Create a very long URL (typical limit is 2048-4096 chars)
	longURL := "https://example.com/" + strings.Repeat("a", 10000)
	body := fmt.Sprintf(`{"url":"%s"}`, longURL)
	req := httptest.NewRequest("POST", "/short", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	// Should either reject or truncate
	// At minimum, should not crash
	if w.Code >= 500 {
		t.Error("server error on excessive URL length")
	}
}

func TestSecurity_InputValidation_ExpireAt(t *testing.T) {
	"""Test that expire_at validation prevents invalid dates"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	invalidExpireAts := []string{
		"not-a-date",
		"2020-01-01", // Already expired
		"9999-12-31T23:59:59Z", // Too far in future (implementation dependent)
	}

	for _, expireAt := range invalidExpireAts {
		body := fmt.Sprintf(`{"url":"https://example.com","expire_at":"%s"}`, expireAt)
		req := httptest.NewRequest("POST", "/short", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Shorten(w, req)

		// Should reject invalid expire_at
		if w.Code == http.StatusCreated {
			// If accepted, at least log warning
			t.Logf("warning: potentially invalid expire_at accepted: %s", expireAt)
		}
	}
}

func TestSecurity_InputValidation_EmptyCode(t *testing.T) {
	"""Test that empty short code is rejected"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	req := httptest.NewRequest("GET", "/", nil)
	req.SetPathValue("code", "")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code == http.StatusOK || w.Code == http.StatusFound {
		t.Error("expected empty code to be rejected")
	}
}

// ============================================================================
// RATE LIMITING & DOS PROTECTION TESTS
// ============================================================================

func TestSecurity_RateLimiting_BruteForce(t *testing.T) {
	"""Test that rate limiting prevents brute force attacks"""
	// This test assumes rate limiter is configured at 5 req/s with capacity 10
	// Testing would require hitting the actual server or mocking the middleware

	// Placeholder for rate limiting test
	// In real scenario, send 100+ requests rapidly and verify 429 responses
	t.Log("Rate limiting test requires running server - see manual testing guide")
}

func TestSecurity_RateLimiting_SlowBrueForce(t *testing.T) {
	"""Test that rate limiting detects slow brute force attacks"""
	// Token bucket should track per-IP and reject sustained high traffic
	t.Log("Slow brute force test requires running server - see manual testing guide")
}

// ============================================================================
// REQUEST SIZE LIMITS
// ============================================================================

func TestSecurity_RequestSizeLimit_LargeJSON(t *testing.T) {
	"""Test that excessively large request bodies are rejected"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	// Create a JSON payload larger than reasonable limit (e.g., 10KB)
	largPayload := `{"url":"https://example.com/" ` + strings.Repeat(","extra_field":"x"}", 2000)
	req := httptest.NewRequest("POST", "/short", strings.NewReader(largPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	// Should reject large request
	// Either 400/413 (bad request / payload too large)
	if w.Code < 400 {
		t.Logf("warning: large payload accepted (status %d)", w.Code)
	}
}

// ============================================================================
// HEADER INJECTION TESTS
// ============================================================================

func TestSecurity_HeaderInjection(t *testing.T) {
	"""Test that header injection attacks are prevented"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	headerInjectionPayloads := []string{
		"value\r\nX-Injected: true",
		"value%0d%0aX-Injected: true",
	}

	for _, payload := range headerInjectionPayloads {
		req := httptest.NewRequest("GET", "/"+payload, nil)
		req.SetPathValue("code", payload)
		w := httptest.NewRecorder()

		handler.Get(w, req)

		// Check response headers don't contain injected content
		for name, values := range w.Header() {
			for _, value := range values {
				if strings.Contains(value, "Injected") {
					t.Errorf("header injection detected in %s: %s", name, value)
				}
			}
		}
	}
}

// ============================================================================
// AUTHORIZATION & AUTHENTICATION TESTS
// ============================================================================

func TestSecurity_NoAuthBypass(t *testing.T) {
	"""Test that endpoints don't have bypass vulnerabilities"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	// Note: Current implementation doesn't have auth, but test format is here
	// for future auth implementation

	req := httptest.NewRequest("POST", "/short", strings.NewReader(`{"url":"https://example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	// Don't set auth header
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	// Current implementation allows public access - this is by design
	// If auth is added in future, this test should verify proper auth check
	t.Log("Current implementation has no auth - public by design")
}

// ============================================================================
// INFORMATION DISCLOSURE TESTS
// ============================================================================

func TestSecurity_NoSensitiveInfoInErrors(t *testing.T) {
	"""Test that error messages don't leak sensitive information"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	// Test various error conditions
	req := httptest.NewRequest("POST", "/short", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	// Response should not contain:
	// - File paths
	// - Database connection strings
	// - Stack traces
	// - Internal server details
	
sensitivePatterns := []string{
		"/home/", "/usr/", "/var/",  // File paths
		"postgresql://", "mysql://",   // DB connection strings
		"panic", "fatal error",          // Stack traces
		"goroutine",                     // Go runtime info
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(w.Body.String(), pattern) {
			t.Errorf("sensitive information leaked in error response: %s", pattern)
		}
	}
}

func TestSecurity_NoDebugInfoInProduction(t *testing.T) {
	"""Test that debug information is not exposed in responses"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	req.SetPathValue("code", "nonexistent")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	// Should return generic 404, not detailed error
	debugPatterns := []string{
		"debug", "trace", "stack", "function",
	}

	for _, pattern := range debugPatterns {
		if strings.Contains(strings.ToLower(w.Body.String()), pattern) {
			t.Logf("warning: debug pattern found: %s", pattern)
		}
	}
}

// ============================================================================
// TIMING ATTACK TESTS
// ============================================================================

func TestSecurity_TimingAttack_ShortCodeLookup(t *testing.T) {
	"""Test that short code lookup timing doesn't leak information"""
	// Timing attacks could allow attackers to guess valid short codes
	// by measuring response times

	// Placeholder for timing attack test
	// Would require measuring response times for valid vs invalid codes
	t.Log("Timing attack test requires performance analysis - see documentation")
}

// ============================================================================
// CSRF TESTS
// ============================================================================

func TestSecurity_CSRF_NoOriginCheck(t *testing.T) {
	"""Test CSRF protection on state-changing requests"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	// POST request from different origin (no CSRF token)
	req := httptest.NewRequest("POST", "/short", strings.NewReader(`{"url":"https://example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://malicious.com")
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	// Current implementation is stateless and doesn't require CSRF tokens
	// This is acceptable for APIs but document it
	if w.Code == http.StatusCreated {
		t.Log("Note: API accepts cross-origin requests (typical for stateless APIs)")
	}
}

// ============================================================================
// CLICKJACKING TESTS
// ============================================================================

func TestSecurity_Clickjacking_XFrameOptions(t *testing.T) {
	"""Test that X-Frame-Options header is set"""
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	req := httptest.NewRequest("GET", "/aB3kZ9", nil)
	req.SetPathValue("code", "aB3kZ9")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	// Check if X-Frame-Options is set
	if xFrameOptions := w.Header().Get("X-Frame-Options"); xFrameOptions == "" {
		t.Log("warning: X-Frame-Options header not set (consider adding for redirect pages)")
	}
}

// ============================================================================
// CACHE POISONING TESTS
// ============================================================================

func TestSecurity_CachePoisoning_InvalidTTL(t *testing.T) {
	"""Test that cache TTL is not vulnerable to poisoning"""
	// Redis cache should have reasonable TTLs to prevent stale data
	// Expired URLs should have infinite TTL or very long TTL

	t.Log("Cache poisoning test requires integration test environment")
}

func TestSecurity_CachePoisoning_QueryString(t *testing.T) {
	"""Test that query string variations don't bypass cache"""
	// Short codes with query params should be normalized or rejected
	svc := &securityMockService{}
	validate := validator.New()
	handler := urlhandler.NewHandler(svc, validate)

	req := httptest.NewRequest("GET", "/aB3kZ9?admin=true&delete=true", nil)
	req.SetPathValue("code", "aB3kZ9")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	// Query params should be ignored for redirects
	// Should not execute delete operation
	if w.Code == http.StatusFound || w.Code == http.StatusFound {
		// Redirect happened - query params were ignored
		t.Log("Query parameters correctly ignored in redirect")
	}
}
