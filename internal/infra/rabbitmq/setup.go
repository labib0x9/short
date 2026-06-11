package rabbitmq

import (
	"fmt"
	"log/slog"

	"github.com/labib0x9/short/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	Queue = "analytics.queue"
)

type RabbitMQ struct {
	conn *amqp.Connection
}

func NewRabbitMQ(cnf *config.RabbitMq) *RabbitMQ {
	url := fmt.Sprintf("amqp://%s:%s@%s/", cnf.User, cnf.Pass, cnf.Addr)
	conn, err := amqp.Dial(url)
	if err != nil {
		panic(fmt.Errorf("rabbitmq dial: %w", err))
		return nil
	}

	r := &RabbitMQ{
		conn: conn,
	}

	if err := r.setup(); err != nil {
		conn.Close()
		panic(err)
		return nil
	}

	slog.Info("rabbitmq connected")

	return r
}

func (r *RabbitMQ) setup() error {
	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("setup channel: %w", err)
	}
	defer ch.Close()

	// dead letter queues

	_, err = ch.QueueDeclare(
		"analytics.queue.dead",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare email dlq: %w", err)
	}

	// main queues

	_, err = ch.QueueDeclare(
		Queue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": "email.queue.dead",
		},
	)
	if err != nil {
		return fmt.Errorf("declare email queue: %w", err)
	}

	return nil
}

func (r *RabbitMQ) Channel() (*amqp.Channel, error) {
	return r.conn.Channel()
}

func (r *RabbitMQ) Close() error {
	if r.conn != nil && !r.conn.IsClosed() {
		return r.conn.Close()
	}

	return nil
}
