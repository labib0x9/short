package url

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/labib0x9/short/internal/domain/url"
)

// TestService_Shorten_Success verifies the happy path: no existing short
// code, no expiry, repo Create succeeds. This is the baseline "Create URL"
// contract every other Shorten test is a variation of.
//
// Setup: urlRepo.GetByShortCode -> (nil, nil) [no collision], Create -> nil.
// Input: a normal https URL, no expiry.
// Expected output: ShortenResult{Msg: "success", Code: <8 chars>, ShortUrl:
// prefix+code, ExpireAt: nil}, err == nil.
// Expected side effects: exactly one Create call with the generated code.
func TestService_Shorten_Success(t *testing.T) {
	repo := &mockUrlRepo{}
	svc := newTestService(repo, nil, nil, nil, nil)

	got, err := svc.Shorten(context.Background(), "https://example.com/very/long/path", nil, "test-agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Msg != "success" {
		t.Errorf("Msg = %q, want %q", got.Msg, "success")
	}
	if len(got.Code) != 8 {
		t.Errorf("Code length = %d, want 8, got %q", len(got.Code), got.Code)
	}
	if got.ShortUrl != "http://localhost:3000/"+got.Code {
		t.Errorf("ShortUrl = %q, want prefix+code", got.ShortUrl)
	}
	if got.ExpireAt != nil {
		t.Errorf("ExpireAt = %v, want nil", got.ExpireAt)
	}
	if repo.CreateCalls != 1 {
		t.Errorf("Create called %d times, want 1", repo.CreateCalls)
	}
}

// TestService_Shorten_WithFutureExpiry verifies expiry is round-tripped
// through to both the stored entity and the returned result unchanged.
func TestService_Shorten_WithFutureExpiry(t *testing.T) {
	repo := &mockUrlRepo{}
	svc := newTestService(repo, nil, nil, nil, nil)

	expireAt := time.Now().Add(24 * time.Hour)
	var createdWith url.Url
	repo.CreateFn = func(ctx context.Context, u url.Url) error {
		createdWith = u
		return nil
	}

	got, err := svc.Shorten(context.Background(), "https://example.com", &expireAt, "test-agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ExpireAt == nil || !got.ExpireAt.Equal(expireAt) {
		t.Errorf("ExpireAt = %v, want %v", got.ExpireAt, expireAt)
	}
	if createdWith.ExpireAt == nil || !createdWith.ExpireAt.Equal(expireAt) {
		t.Errorf("repo received ExpireAt = %v, want %v", createdWith.ExpireAt, expireAt)
	}
}

// TestService_Shorten_ExpiredAtCreation is a business-rule/invariant test:
// creating a link whose expiry is already in the past must be rejected
// before it ever reaches the repository, otherwise the system would create
// permanently-dead links.
func TestService_Shorten_ExpiredAtCreation(t *testing.T) {
	repo := &mockUrlRepo{}
	svc := newTestService(repo, nil, nil, nil, nil)

	past := time.Now().Add(-24 * time.Hour)

	_, err := svc.Shorten(context.Background(), "https://example.com", &past, "test-agent")
	if !errors.Is(err, url.ErrShortCodeExpired) {
		t.Fatalf("err = %v, want %v", err, url.ErrShortCodeExpired)
	}
	if repo.CreateCalls != 0 {
		t.Errorf("Create called %d times, want 0 (must not persist an already-expired link)", repo.CreateCalls)
	}
}

// TestService_Shorten_CollisionDetected: GetByShortCode returning an
// existing, non-nil Url must short-circuit with ErrShortCodeCollision and
// never call Create. Documents that this service does NOT auto-retry with a
// freshly generated code - collision handling is the caller's job.
func TestService_Shorten_CollisionDetected(t *testing.T) {
	repo := &mockUrlRepo{
		GetByShortCodeFn: func(ctx context.Context, shortCode string) (*url.Url, error) {
			return &url.Url{ShortURL: shortCode}, nil
		},
	}
	svc := newTestService(repo, nil, nil, nil, nil)

	_, err := svc.Shorten(context.Background(), "https://example.com", nil, "test-agent")
	if !errors.Is(err, url.ErrShortCodeCollision) {
		t.Fatalf("err = %v, want %v", err, url.ErrShortCodeCollision)
	}
	if repo.CreateCalls != 0 {
		t.Errorf("Create called %d times, want 0 on collision", repo.CreateCalls)
	}
}

// TestService_Shorten_NoRowsIsNotAnError verifies sql.ErrNoRows from the
// existence check is treated as "no collision", not as a failure - this is
// the expected/common path for GetByShortCode on a fresh, unused code.
func TestService_Shorten_NoRowsIsNotAnError(t *testing.T) {
	repo := &mockUrlRepo{
		GetByShortCodeFn: func(ctx context.Context, shortCode string) (*url.Url, error) {
			return nil, sql.ErrNoRows
		},
	}
	svc := newTestService(repo, nil, nil, nil, nil)

	_, err := svc.Shorten(context.Background(), "https://example.com", nil, "test-agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.CreateCalls != 1 {
		t.Errorf("Create called %d times, want 1", repo.CreateCalls)
	}
}

// TestService_Shorten_RepositoryErrors is a table-driven test covering
// repository failures at each stage of Shorten and asserting the error is
// surfaced unchanged (no swallowing / no wrapping that would break
// errors.Is downstream) and that side effects stop at the failing step.
func TestService_Shorten_RepositoryErrors(t *testing.T) {
	dbErr := errors.New("connection refused")
	createErr := errors.New("insert failed")

	tests := []struct {
		name          string
		getByCodeFn   func(ctx context.Context, shortCode string) (*url.Url, error)
		createFn      func(ctx context.Context, u url.Url) error
		wantErr       error
		wantCreateCnt int
	}{
		{
			name: "GetByShortCode non-NoRows error is returned as-is and Create is never called",
			getByCodeFn: func(ctx context.Context, shortCode string) (*url.Url, error) {
				return nil, dbErr
			},
			wantErr:       dbErr,
			wantCreateCnt: 0,
		},
		{
			name: "Create error is returned as-is",
			getByCodeFn: func(ctx context.Context, shortCode string) (*url.Url, error) {
				return nil, nil
			},
			createFn: func(ctx context.Context, u url.Url) error {
				return createErr
			},
			wantErr:       createErr,
			wantCreateCnt: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUrlRepo{
				GetByShortCodeFn: tt.getByCodeFn,
				CreateFn:         tt.createFn,
			}
			svc := newTestService(repo, nil, nil, nil, nil)

			_, err := svc.Shorten(context.Background(), "https://example.com", nil, "test-agent")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if repo.CreateCalls != tt.wantCreateCnt {
				t.Errorf("Create called %d times, want %d", repo.CreateCalls, tt.wantCreateCnt)
			}
		})
	}
}

// TestService_Shorten_ContextCancellationPropagates verifies a cancelled
// context is passed through to the repository call rather than being
// silently ignored - Shorten has no explicit ctx.Done() check of its own, so
// this pins down that cancellation enforcement is (correctly) delegated to
// the repository/driver layer, and that Shorten doesn't swallow that error.
func TestService_Shorten_ContextCancellationPropagates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := &mockUrlRepo{
		GetByShortCodeFn: func(ctx context.Context, shortCode string) (*url.Url, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return nil, nil
		},
	}
	svc := newTestService(repo, nil, nil, nil, nil)

	_, err := svc.Shorten(ctx, "https://example.com", nil, "test-agent")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

// TestService_Shorten_IsNotIdempotent documents (rather than "fixes") an
// important business characteristic: because GetShortUrl mixes in a random
// seed, calling Shorten twice with the identical URL and no existing
// collision produces two distinct records with two distinct codes - it is
// NOT idempotent on the input URL. Anyone building a client that expects
// "shorten the same URL twice -> same short code" needs to know this
// up front; if the implementation ever changes to be idempotent, this test
// will fail loudly and force the docs/README to be updated too.
func TestService_Shorten_IsNotIdempotent(t *testing.T) {
	repo := &mockUrlRepo{}
	svc := newTestService(repo, nil, nil, nil, nil)

	const longUrl = "https://example.com/repeatedly/shortened"

	first, err := svc.Shorten(context.Background(), longUrl, nil, "test-agent")
	if err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}
	second, err := svc.Shorten(context.Background(), longUrl, nil, "test-agent")
	if err != nil {
		t.Fatalf("second call: unexpected error: %v", err)
	}

	if first.Code == second.Code {
		t.Fatalf("expected distinct codes across calls (non-idempotent by design), got the same code %q twice", first.Code)
	}
	if repo.CreateCalls != 2 {
		t.Errorf("Create called %d times, want 2 (each call persists its own record)", repo.CreateCalls)
	}
}

// TestService_Shorten_ConcurrentCalls fires many concurrent Shorten calls at
// the service to verify there's no data race in the service itself (run
// with -race) and that every call either succeeds or fails independently -
// a panic or shared mutable state bug here would surface as a race
// detector failure or a lost/duplicated Create call count.
func TestService_Shorten_ConcurrentCalls(t *testing.T) {
	repo := &mockUrlRepo{}
	svc := newTestService(repo, nil, nil, nil, nil)

	const n = 50
	errCh := make(chan error, n)
	done := make(chan struct{})
	var started int
	for i := 0; i < n; i++ {
		go func(i int) {
			started++
			_, err := svc.Shorten(context.Background(), "https://example.com/concurrent", nil, "test-agent")
			errCh <- err
		}(i)
	}
	go func() {
		for i := 0; i < n; i++ {
			<-errCh
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for concurrent Shorten calls to finish")
	}

	if repo.callCount(func() int { return repo.CreateCalls }) != n {
		t.Errorf("Create called %d times, want %d", repo.CreateCalls, n)
	}
}
