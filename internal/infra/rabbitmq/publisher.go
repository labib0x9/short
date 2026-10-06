package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	queueport "github.com/labib0x9/short/internal/port/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	defaultExchange = ""
)

// queue = Queue name to publish msg
func (r *rabbitMQ) publish(ctx context.Context, queue string, msg any) error {
	ch, err := r.channel()
	if err != nil {
		return fmt.Errorf("%v: %w", queueport.ErrOpeningChannel, err)
	}
	defer ch.Close()

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("%v: %w", queueport.ErrMessageEncodingFailed, err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(
		ctx,
		defaultExchange,
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)

	if err != nil {
		return fmt.Errorf("%v: %w", queueport.ErrPublishMessageFailed, err)
	}

	slog.Info("Message Published", "Queue:", queue)
	return nil
}

func (r *rabbitMQ) PublishAnalytics(ctx context.Context, msg queueport.ClickEvent) error {
	return r.publish(ctx, AnalyticQueue, msg)
}
