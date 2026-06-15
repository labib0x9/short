package rabbitmq_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/labib0x9/short/config"
	"github.com/labib0x9/short/internal/domain/queue"
	"github.com/labib0x9/short/internal/infra/rabbitmq"
	"github.com/stretchr/testify/assert"
)

func TestRabbitMQPublishAndConsume(t *testing.T) {
	cnf := &config.RabbitMq{
		User: "guest",
		Pass: "guest",
		Addr: "localhost:5672",
	}

	q := rabbitmq.NewRabbitMQ(cnf)
	defer q.Close()

	ctx := context.Background()

	// Publish a test event
	event := queue.ClickEvent{
		ShortCode: "abc123",
		ClickedAt: time.Now(),
		Referer:   "https://google.com",
		UserAgent: "Mozilla/5.0 (Linux; Android 14; SM-F956U) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/80.0.3987.119 Mobile Safari/537.36",
		IP:        "127.0.0.1",
	}

	err := q.PublishAnalytics(ctx, event)
	assert.NoError(t, err, "publish should succeed")

	// Consume and verify
	msgs, err := q.ConsumeAnalytics(ctx, "test-consumer", 1)
	assert.NoError(t, err)

	select {
	case d := <-msgs:
		var got queue.ClickEvent
		err = json.Unmarshal(d.Body, &got)
		assert.NoError(t, err)
		assert.Equal(t, event.ShortCode, got.ShortCode)
		d.Ack(false)
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}
