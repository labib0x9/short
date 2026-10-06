package queue

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Queue interface {
	PublishAnalytics(ctx context.Context, msg ClickEvent) error
	ConsumeAnalytics(ctx context.Context, name string, concurrency int) (<-chan amqp.Delivery, error)
	Close() error
	CloseConsumerChannel(name string) error
}

type Delivery interface {
	Body() []byte
	Ack(multiple bool) error
	Nack(multiple bool, requeue bool) error
	Reject(requeue bool) error
}
