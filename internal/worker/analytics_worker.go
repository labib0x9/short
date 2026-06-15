package worker

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/labib0x9/short/internal/app/url"
	"github.com/labib0x9/short/internal/domain/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

type deliveryWrapper struct {
	d amqp.Delivery
}

func (w *deliveryWrapper) Body() []byte {
	return w.d.Body
}

func (w *deliveryWrapper) Ack(multiple bool) error {
	return w.d.Ack(multiple)
}

func (w *deliveryWrapper) Nack(multiple bool, requeue bool) error {
	return w.d.Nack(multiple, requeue)
}

func (w *deliveryWrapper) Reject(requeue bool) error {
	return w.d.Reject(requeue)
}

type Worker struct {
	queue      queue.Queue
	srv        url.Service
	maxRetries int
}

func NewWorker(
	queue queue.Queue,
	srv url.Service,
) *Worker {
	return &Worker{
		queue:      queue,
		srv:        srv,
		maxRetries: 2,
	}
}

// name = worker/consumer identifier
func (w *Worker) Run(ctx context.Context, name string, concurrency int) error {
	msgs, err := w.queue.ConsumeAnalytics(ctx, name, concurrency)
	if err != nil {
		return err
	}
	defer w.queue.CloseConsumerChannel(name)

	slog.Info("Analytics worker started", "Concurrency", concurrency)

	sem := make(chan struct{}, concurrency)
	for {
		select {
		case <-ctx.Done():
			slog.Info("Analytics worker shutting down")
			return nil
		case d, ok := <-msgs:
			if !ok {
				return queue.ErrConsumerChannelClosed
			}
			sem <- struct{}{}
			go func(d amqp.Delivery) {
				defer func() {
					<-sem
				}()
				w.handle(ctx, &deliveryWrapper{d: d})
			}(d)
		}
	}
}

func (w *Worker) handle(ctx context.Context, d queue.Delivery) {
	var msg queue.ClickEvent
	err := json.Unmarshal(d.Body(), &msg)
	if err != nil {
		slog.Error("handler() Json read failed", "error", err)
		d.Nack(false, false)
		return
	}

	err = w.srv.Save(ctx, msg)

	if err != nil {
		if msg.Retries < w.maxRetries {
			msg.Retries++
			if err := d.Nack(false, true); err != nil {
				slog.Error("handle() Requeue failed", "code", msg.ShortCode, "error", err)
			}
			return
		}

		if err := d.Nack(false, false); err != nil {
			slog.Error("handle() Queue to DLQ failed", "code", msg.ShortCode, "error", err)
		}
		return
	}

	if err = d.Ack(false); err != nil {
		slog.Error("handle() Ack send failed", "code", msg.ShortCode, "error", err)
		return
	}

	slog.Info("analytics-worker processed successfully", "code", msg.ShortCode)
}
