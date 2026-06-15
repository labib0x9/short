package rabbitmq

import (
	"fmt"
	"log/slog"

	"github.com/labib0x9/short/config"
	"github.com/labib0x9/short/internal/domain/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	AnalyticQueue     = "analytics.queue"
	AnalyticQueueDead = "analytics.queue.dead"
)

type rabbitMQ struct {
	conn       *amqp.Connection
	consumerCh map[string]*amqp.Channel // for each consumer a dedicated channel
}

func NewRabbitMQ(cnf *config.RabbitMq) queue.Queue {
	url := fmt.Sprintf("amqp://%s:%s@%s/", cnf.User, cnf.Pass, cnf.Addr)
	conn, err := amqp.Dial(url)
	if err != nil {
		panic(fmt.Errorf("rabbitmq dial: %w, url=%s", err, url))
	}

	r := rabbitMQ{
		conn:       conn,
		consumerCh: make(map[string]*amqp.Channel),
	}

	if err := r.setup(); err != nil {
		conn.Close()
		panic(err)
	}

	slog.Info("rabbitMq connection complete")

	return &r
}

func (r *rabbitMQ) setup() error {
	ch, err := r.channel()
	if err != nil {
		return fmt.Errorf("%v: %w", queue.ErrOpeningChannel, err)
	}
	defer ch.Close()

	err = r.declareAnalyticQueueDead(ch)
	if err != nil {
		return fmt.Errorf("%v: %w", queue.ErrDeclareQueue, err)
	}

	err = r.declareAnalyticQueue(ch)
	if err != nil {
		return fmt.Errorf("%v: %w", queue.ErrDeclareQueue, err)
	}

	return nil
}

func (r *rabbitMQ) declareAnalyticQueue(ch *amqp.Channel) error {
	_, err := ch.QueueDeclare(
		AnalyticQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": AnalyticQueueDead,
		},
	)
	return err
}

func (r *rabbitMQ) declareAnalyticQueueDead(ch *amqp.Channel) error {
	_, err := ch.QueueDeclare(
		AnalyticQueueDead,
		true,
		false,
		false,
		false,
		nil,
	)
	return err
}

func (r *rabbitMQ) channel() (*amqp.Channel, error) {
	return r.conn.Channel()
}

func (r *rabbitMQ) CloseConsumerChannel(name string) error {
	err := r.consumerCh[name].Close()
	if err != nil {
		return err
	}
	delete(r.consumerCh, name)
	return nil
}

func (r *rabbitMQ) Close() error {

	for _, ch := range r.consumerCh {
		if ch != nil {
			ch.Close()
		}
	}

	if r.conn != nil && !r.conn.IsClosed() {
		return r.conn.Close()
	}
	return nil
}
