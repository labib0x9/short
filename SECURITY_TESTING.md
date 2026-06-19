# Security Testing Guide

This guide covers security testing for the **short** URL shortener project.

---

## 🔒 Security Test Categories

### 1. **SQL Injection Tests**
Prevents attackers from executing arbitrary SQL commands

```bash
go test -v -run TestSecurity_SQLInjection ./tests
```

**Payloads Tested:**
- `'; DROP TABLE urls; --`
- `1' OR '1'='1`
- `' UNION SELECT * FROM users --`
- `1; DELETE FROM urls; --`

**What's Verified:**
- ✅ Parameterized queries prevent injection
- ✅ SQL errors don't leak in responses
- ✅ Malicious codes return 404 or 400, not 500

---

### 2. **XSS (Cross-Site Scripting) Tests**
Prevents malicious scripts from being executed in browsers

```bash
go test -v -run TestSecurity_XSS ./tests
```

**Payloads Tested:**
- `javascript:alert('XSS')`
- `<script>alert('XSS')</script>`
- `<img src=x onerror=alert('XSS')>`

**What's Verified:**
- ✅ Invalid URL schemes rejected
- ✅ HTML/JS in responses is escaped
- ✅ Error messages don't echo unescaped input

---

### 3. **Input Validation Tests**
Ensures all inputs are properly validated

```bash
go test -v -run TestSecurity_InputValidation ./tests
```

**Tests Include:**
- ✅ Invalid URL formats rejected
- ✅ Unsupported schemes blocked (ftp://, file://, gopher://)
- ✅ Excessively long URLs rejected or truncated
- ✅ Invalid expire_at dates rejected
- ✅ Empty short codes rejected

**Invalid URLs Tested:**
- `not-a-url`
- `ftp://example.com`
- `javascript:void(0)`
- `file:///etc/passwd`
- Empty string

---

### 4. **Rate Limiting & DOS Protection**
Prevents denial-of-service attacks

```bash
go test -v -run TestSecurity_RateLimiting ./tests
```

**Verification:**
- ✅ 5 req/s rate limit per IP
- ✅ Capacity of 10 tokens (burst handling)
- ✅ 429 Too Many Requests response on breach
- ✅ Retry-After header present

**Manual Testing:**
```bash
# Send 100 requests in 10 seconds (should be throttled)
for i in {1..100}; do
  curl -i http://localhost:3000/aB3kZ9 &
  sleep 0.1
done

# Should see:
# HTTP/1.1 200 OK (first 10)
# HTTP/1.1 429 Too Many Requests (rest)
```

---

### 5. **Request Size Limits**
Prevents memory exhaustion attacks

```bash
go test -v -run TestSecurity_RequestSizeLimit ./tests
```

**What's Tested:**
- ✅ Large JSON payloads rejected
- ✅ Malformed JSON doesn't crash server
- ✅ Memory usage stays bounded

**Manual Test:**
```bash
# Create 50MB JSON file
dd if=/dev/zero bs=1M count=50 | base64 > large_payload.txt

# Try to upload
curl -X POST http://localhost:3000/short \
  -H "Content-Type: application/json" \
  --data-binary @large_payload.txt

# Should get 413 Payload Too Large or timeout
```

---

### 6. **Header Injection Tests**
Prevents HTTP response splitting and header injection

```bash
go test -v -run TestSecurity_HeaderInjection ./tests
```

**Payloads Tested:**
- `value\r\nX-Injected: true`
- `value%0d%0aX-Injected: true` (URL encoded CRLF)

**What's Verified:**
- ✅ CRLF characters are rejected or sanitized
- ✅ No injected headers in response
- ✅ Response splitting prevented

---

### 7. **Information Disclosure Tests**
Ensures sensitive information isn't leaked in errors

```bash
go test -v -run TestSecurity_NoSensitiveInfo ./tests
```

**Information That Should NOT Leak:**
- ❌ File paths (`/home/`, `/var/`, `/usr/`)
- ❌ Database connection strings
- ❌ Stack traces
- ❌ Go runtime info (goroutines, panic messages)
- ❌ Internal server details

**Example - Dangerous Error:**
```json
{
  "error": "database connection failed",
  "details": "failed to connect to postgresql://short:secret@localhost:5432/short"
}
```

**Example - Safe Error:**
```json
{
  "error": "internal server error"
}
```

---

### 8. **CSRF (Cross-Site Request Forgery) Tests**

```bash
go test -v -run TestSecurity_CSRF ./tests
```

**Current Status:**
- ✅ API is stateless (doesn't use sessions/cookies)
- ✅ No CSRF tokens needed (stateless design)
- ✅ Safe from CSRF attacks by design

**Note:** If session-based auth is added in future, CSRF tokens must be implemented.

---

### 9. **Clickjacking Protection**

```bash
go test -v -run TestSecurity_Clickjacking ./tests
```

**Verification:**
- ⚠️ Currently checks for X-Frame-Options header
- ℹ️ Consider adding for static redirect pages

**Recommended Headers:**
```
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
```

---

### 10. **Timing Attack Tests**

```bash
go test -v -run TestSecurity_TimingAttack ./tests
```

**What Could Be Exploited:**
- Attackers could measure response time differences
- Valid codes might take longer to look up than invalid ones
- Could allow brute-forcing valid short codes

**Mitigation:**
- Use constant-time comparison for sensitive operations
- Add random jitter to response times
- Rate limiting (already implemented)

---

## 🚀 Running All Security Tests

```bash
# Run all security tests
go test -v -tags=security ./tests

# Run specific category
go test -v -run TestSecurity_SQLInjection ./tests

# Run with coverage
go test -v -tags=security -coverprofile=security.out ./tests
go tool cover -html=security.out
```

---

## 🔍 Manual Security Testing Checklist

```bash
# Start server and services
docker-compose -f docker-compose.test.yml up -d
go run ./cmd/migration/main.go -setup
go run ./cmd/migration/main.go -up
go run ./cmd/short/main.go &
```

### SQL Injection Tests
```bash
# Test with SQL injection payload
curl "http://localhost:3000/'; DROP TABLE urls; --/stat"
# Expected: 404 Not Found, no SQL error

curl "http://localhost:3000/1' OR '1'='1/stat"
# Expected: 404 Not Found
```

### XSS Tests
```bash
# Test with XSS payload in URL
curl -X POST http://localhost:3000/short \
  -H "Content-Type: application/json" \
  -d '{"url":"javascript:alert(1)"}'
# Expected: 400 Bad Request (invalid URL scheme)

curl -X POST http://localhost:3000/short \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com<script>alert(1)</script>"}'
# Expected: 400 Bad Request or accepted but escaped in response
```

### Rate Limiting Tests
```bash
# Send 50 requests rapidly (limit is 5 req/s, capacity 10)
for i in {1..50}; do
  curl -i http://localhost:3000/aB3kZ9 > /dev/null 2>&1 &
done

# Monitor response codes
# Expected:
# - First 10 requests: 302 (redirect)
# - Remaining requests: 429 (too many requests)
```

### Request Size Limit Tests
```bash
# Create large JSON payload
python3 -c "import json; print(json.dumps({'url': 'https://example.com/' + 'x'*10000}))" > large.json

# Try to send
curl -X POST http://localhost:3000/short \
  -H "Content-Type: application/json" \
  --data-binary @large.json
# Expected: 413 Payload Too Large or timeout
```

### Error Message Disclosure Tests
```bash
# Trigger error by accessing non-existent code
curl http://localhost:3000/nonexistent/stat -v

# Check error message
# ✅ Safe: {"error":"not found"}
# ❌ Unsafe: Full stack trace or file paths

# Trigger JSON parse error
curl -X POST http://localhost:3000/short \
  -H "Content-Type: application/json" \
  -d '{invalid json}'
# Check response doesn't contain sensitive info
```

### HTTPS/TLS Tests
```bash
# Verify server can be accessed over HTTPS in production
# (if deployed with TLS)
curl -v https://localhost:3000/aB3kZ9

# Check for security headers
curl -i https://localhost:3000/aB3kZ9 | grep -E "X-Frame-Options|X-Content-Type-Options|X-XSS-Protection"
```

---

## 🛡️ Security Best Practices

### Currently Implemented ✅

1. **Parameterized Queries**
   - All database queries use parameterized queries (sqlx)
   - Prevents SQL injection

2. **Input Validation**
   - URL format validation (go-playground/validator)
   - URL scheme validation (http/https only)
   - Empty field checks

3. **Rate Limiting**
   - Token bucket algorithm in Redis
   - 5 req/s per IP with capacity 10
   - Atomic Lua script prevents race conditions

4. **Error Handling**
   - Generic error messages (no SQL details leaked)
   - No stack traces in responses

5. **Graceful Shutdown**
   - SIGINT/SIGTERM handling
   - Allows connections to drain

### Recommended Additions ⚠️

1. **HTTPS/TLS**
   ```go
   // Add to server startup
   server.ListenAndServeTLS("cert.pem", "key.pem")
   ```

2. **Security Headers**
   ```go
   func securityHeaders(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           w.Header().Set("X-Frame-Options", "DENY")
           w.Header().Set("X-Content-Type-Options", "nosniff")
           w.Header().Set("X-XSS-Protection", "1; mode=block")
           w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
           next.ServeHTTP(w, r)
       })
   }
   ```

3. **Content Security Policy**
   ```go
   w.Header().Set("Content-Security-Policy", "default-src 'none'")
   ```

4. **Input Size Limits**
   ```go
   r.Body = http.MaxBytesReader(w, r.Body, 10*1024) // 10KB limit
   ```

5. **Constant-Time Comparison**
   ```go
   import "crypto/subtle"
   if subtle.ConstantTimeCompare([]byte(code), []byte(validCode)) == 0 {
       return ErrNotFound
   }
   ```

6. **Audit Logging**
   - Log all URL creation with IP/timestamp
   - Log suspicious activity (rate limit breaches)
   - Log failed auth attempts (if auth added)

7. **API Rate Limiting Per User**
   - If auth added, limit per user not just IP
   - Prevent account enumeration

---

## 📊 Security Testing Report

| Test Category | Status | Risk Level | Notes |
|---------------|--------|------------|-------|
| SQL Injection | ✅ PASS | Low | Parameterized queries |
| XSS | ✅ PASS | Low | URL validation prevents most payloads |
| Input Validation | ✅ PASS | Low | URL format validated |
| Rate Limiting | ✅ PASS | Low | Token bucket per IP |
| Request Size | ⚠️ WARN | Medium | No hard limit, consider adding |
| Header Injection | ✅ PASS | Low | Go stdlib handles CRLF |
| Info Disclosure | ✅ PASS | Low | Generic error messages |
| CSRF | ✅ PASS | N/A | Stateless API design |
| Clickjacking | ⚠️ WARN | Medium | Consider X-Frame-Options |
| Timing Attacks | ⚠️ WARN | Low | Rate limiting mitigates |

---

## 🔐 Deployment Security Checklist

- [ ] HTTPS/TLS enabled with valid certificate
- [ ] Security headers configured (X-Frame-Options, CSP, etc.)
- [ ] Environment variables not logged
- [ ] Database credentials in .env (not hardcoded)
- [ ] Redis/RabbitMQ require authentication
- [ ] PostgreSQL user has minimal privileges
- [ ] Request logging doesn't include sensitive data
- [ ] Regular security updates (Go, dependencies)
- [ ] OWASP Top 10 remediated
- [ ] Security scanning in CI/CD pipeline

---

## 📚 References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://go.dev/blog/security-blog)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [CWE Top 25](https://cwe.mitre.org/top25/)
