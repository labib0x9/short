package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/labib0x9/short/internal/domain/url"
	"github.com/labib0x9/short/internal/infra/cache"
	"github.com/labib0x9/short/internal/infra/queue"
	"github.com/labib0x9/short/internal/infra/rabbitmq"
	"github.com/mileusna/useragent"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Worker struct {
	client       *rabbitmq.RabbitMQ
	urlRepo      url.UrlRepository
	analysisRepo url.AnalyticsRepository
	cache        cache.CacheRepo
	maxRetries   int
}

func NewWorker(
	client *rabbitmq.RabbitMQ,
	urlRepo url.UrlRepository,
	analysisRepo url.AnalyticsRepository,
	cacheRepo cache.CacheRepo,
) *Worker {
	return &Worker{
		client:       client,
		maxRetries:   2,
		cache:        cacheRepo,
		urlRepo:      urlRepo,
		analysisRepo: analysisRepo,
	}
}

func (w *Worker) Run(ctx context.Context, concurrency int) error {

	// dedicated consumer channel
	ch, err := w.client.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	// limit unacked messages
	err = ch.Qos(concurrency, 0, false)
	if err != nil {
		return fmt.Errorf("qos: %w", err)
	}

	msgs, err := ch.Consume(
		rabbitmq.Queue,
		"analytics-worker",
		false, // auto ack
		false, // exclusive
		false, // no local
		false, // no wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	slog.Info("analytics worker started", "concurrency", concurrency)

	sem := make(chan struct{}, concurrency)
	for {
		select {
		case <-ctx.Done():
			slog.Info("analytic worker shutting down")
			return nil
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("consumer channel closed")
			}
			sem <- struct{}{}
			go func(d amqp.Delivery) {
				defer func() {
					<-sem
				}()
				w.handle(ctx, d)
			}(d)
		}
	}
}

func (w *Worker) handle(ctx context.Context, d amqp.Delivery) {
	slog.Info("Inside Worker Msg Queue")
	var msg queue.ClickEvent
	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		slog.Error("invalid message", "error", err)
		d.Nack(false, false)
		return
	}

	// to-do
	// Get country from ip

	device, browser, os := ParseUserAgent(msg.UserAgent)
	if device == "" || browser == "" || os == "" {
		//
	}

	found, err := w.urlRepo.GetByShortCode(msg.ShortCode)
	if err != nil {
		slog.Error("fetch code failed", "error", err)
	}

	click := url.Click{
		UrlId:      found.Id,
		Referer:    msg.Referer,
		Country:    "",
		DeviceType: device,
		Os:         os,
		Browser:    browser,
		ClickedAt:  msg.ClickedAt,
	}

	err = w.analysisRepo.Create(click)
	if err != nil {
		slog.Error("click insert failed", "error", err)
	}

	if err == nil {
		err = w.urlRepo.Update(found.Id, msg.ClickedAt)
		if err != nil {
			slog.Error("url update failed", "error", err)
		}
	}

	if err != nil {
		// retry
		if msg.Retries < w.maxRetries {
			msg.Retries++
			err := d.Nack(false, true)
			if err != nil {
				slog.Error("nack retry failed", "error", err)
			}
			return
		}

		// dead-letter
		err := d.Nack(false, false)
		if err != nil {
			slog.Error("nack dead-letter failed", "error", err)
		}
		return
	}

	err = d.Ack(false)
	if err != nil {
		slog.Error("ack failed", "error", err)
		return
	}

	slog.Info("analytics processed successfully", "short-code", msg.ShortCode)
	// return nil
}

// func retryCount(d amqp.Delivery) int {
// 	deaths, ok := d.Headers["x-death"].([]interface{})
// 	if !ok || len(deaths) == 0 {
// 		return 0
// 	}
// 	entry, ok := deaths[0].(amqp.Table)
// 	if !ok {
// 		return 0
// 	}
// 	count, _ := entry["count"].(int64)
// 	return int(count)
// }

func ParseUserAgent(agent string) (device string, browser string, os string) {
	ua := useragent.Parse(agent)

	switch {
	case ua.Mobile:
		device = "Mobile"
	case ua.Desktop:
		device = "Desktop"
	}

	os = ua.OS

	browser = ua.Name

	switch {
	case device == "":
		device = "unknown"
	case browser == "":
		browser = "unknown"
	case os == "":
		os = "unknown"
	}

	return
}
