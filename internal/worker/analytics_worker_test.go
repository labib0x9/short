package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/labib0x9/short/internal/app/url"
	"github.com/labib0x9/short/internal/domain/queue"
	urldomain "github.com/labib0x9/short/internal/domain/url"
	amqp "github.com/rabbitmq/amqp091-go"
)

// --- Mock: queue.Queue ---
type mockQueue struct {
	msgs chan amqp.Delivery
}

func (m *mockQueue) ConsumeAnalytics(_ context.Context, _ string, _ int) (<-chan amqp.Delivery, error) {
	return m.msgs, nil
}
func (m *mockQueue) PublishAnalytics(_ context.Context, _ queue.ClickEvent) error { return nil }
func (m *mockQueue) Close() error                                                 { return nil }
func (m *mockQueue) CloseConsumerChannel(_ string) error                          { return nil }

// --- Mock: queue.Delivery ---
type mockDelivery struct {
	body    []byte
	acked   bool
	nacked  bool
	requeue bool
}

func (m *mockDelivery) Body() []byte {
	return m.body
}

func (m *mockDelivery) Ack(_ bool) error {
	m.acked = true
	return nil
}
func (m *mockDelivery) Nack(_ bool, requeue bool) error {
	m.nacked = true
	m.requeue = requeue
	return nil
}
func (m *mockDelivery) Reject(_ bool) error { return nil }

// --- Mock: url.Service ---
type mockService struct {
	saveErr   error
	saveCalls int
}

func (m *mockService) Save(ctx context.Context, msg queue.ClickEvent) error {
	m.saveCalls++
	return m.saveErr
}

func (m *mockService) Shorten(ctx context.Context, longUrl string, expireAt *time.Time, userAgent string) (url.ShortenResult, error) {
	return url.ShortenResult{}, nil
}
func (m *mockService) Get(ctx context.Context, code, referer, userAgent, remoteAddr string) (*urldomain.Url, error) {
	return nil, nil
}
func (m *mockService) Analysis(ctx context.Context, code string) (urldomain.Analysis, error) {
	return urldomain.Analysis{}, nil
}

func (m *mockService) DeleteByExpireAt(ctx context.Context) error {
	return nil
}

// --- Helpers ---
func makeDelivery(t *testing.T, event queue.ClickEvent) *mockDelivery {
	t.Helper()
	body, _ := json.Marshal(event)
	return &mockDelivery{body: body}
}

func makeWorker(svc *mockService) *Worker {
	return &Worker{
		queue:      &mockQueue{},
		srv:        svc,
		maxRetries: 2,
	}
}

// --- Tests ---

func TestHandle_Success(t *testing.T) {
	svc := &mockService{}
	w := makeWorker(svc)

	event := queue.ClickEvent{ShortCode: "abc123", ClickedAt: time.Now()}
	d := makeDelivery(t, event)

	w.handle(context.Background(), d)

	if !d.acked {
		t.Error("expected Ack to be called")
	}
	if d.nacked {
		t.Error("expected Nack NOT to be called")
	}
	if svc.saveCalls != 1 {
		t.Errorf("expected Save to be called once, got %d", svc.saveCalls)
	}
}

func TestHandle_InvalidJSON(t *testing.T) {
	w := makeWorker(&mockService{})
	d := &mockDelivery{body: []byte("not-json")}

	w.handle(context.Background(), d)

	if !d.nacked {
		t.Error("expected Nack on bad JSON")
	}
	if d.requeue {
		t.Error("expected requeue=false on bad JSON")
	}
}

func TestHandle_ServiceError_Retries(t *testing.T) {
	svc := &mockService{saveErr: errors.New("db error")}
	w := makeWorker(svc)

	// Retries = 0, below maxRetries(2) → should requeue
	event := queue.ClickEvent{ShortCode: "abc123", Retries: 0}
	d := makeDelivery(t, event)

	w.handle(context.Background(), d)

	if !d.nacked {
		t.Error("expected Nack")
	}
	if !d.requeue {
		t.Error("expected requeue=true when retries < maxRetries")
	}
}

func TestHandle_ServiceError_MaxRetriesExceeded(t *testing.T) {
	svc := &mockService{saveErr: errors.New("db error")}
	w := makeWorker(svc)

	// Retries = 2, equal to maxRetries → should go to DLQ
	event := queue.ClickEvent{ShortCode: "abc123", Retries: 2}
	d := makeDelivery(t, event)

	w.handle(context.Background(), d)

	if !d.nacked {
		t.Error("expected Nack")
	}
	if d.requeue {
		t.Error("expected requeue=false when max retries exceeded → DLQ")
	}
}

func TestRun_ContextCancel(t *testing.T) {
	msgs := make(chan amqp.Delivery)
	q := &mockQueue{msgs: msgs}
	w := &Worker{queue: q, srv: &mockService{}, maxRetries: 2}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- w.Run(ctx, "test-worker", 1)
	}()

	cancel() // trigger shutdown

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil on context cancel, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not shut down in time")
	}
}

func TestRun_ChannelClosed(t *testing.T) {
	msgs := make(chan amqp.Delivery)
	q := &mockQueue{msgs: msgs}
	w := &Worker{queue: q, srv: &mockService{}, maxRetries: 2}

	done := make(chan error)
	go func() {
		done <- w.Run(context.Background(), "test-worker", 1)
	}()

	close(msgs) // simulate broker closing the channel

	select {
	case err := <-done:
		if !errors.Is(err, queue.ErrConsumerChannelClosed) {
			t.Errorf("expected ErrConsumerChannelClosed, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not return on channel close")
	}
}
