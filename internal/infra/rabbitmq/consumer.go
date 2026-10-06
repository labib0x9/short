package rabbitmq

import (
	"context"
	"fmt"

	queueport "github.com/labib0x9/short/internal/port/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

// queue = queue name, from where consumer collects msg
// name = consumer identity
func (r *rabbitMQ) consume(
	ctx context.Context,
	queue string,
	name string,
	concurrency int,
	autoAck bool,
	exclusive bool,
	noLocal bool,
	noWait bool,
	args amqp.Table,
) (<-chan amqp.Delivery, error) {
	ch, err := r.channel()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", queueport.ErrOpeningChannel, err)
	}

	r.consumerCh[name] = ch

	err = ch.Qos(concurrency, 0, false)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", queueport.ErrQoSFailed, err)
	}

	return ch.ConsumeWithContext(
		ctx,
		queue,
		name,
		autoAck,
		exclusive,
		noLocal,
		noWait,
		args,
	)
}

func (r *rabbitMQ) ConsumeAnalytics(ctx context.Context, name string, concurrency int) (<-chan amqp.Delivery, error) {
	return r.consume(ctx, AnalyticQueue, name, concurrency, false, false, false, false, nil)
}
