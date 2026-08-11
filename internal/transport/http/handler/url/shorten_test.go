package url

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labib0x9/short/config"
	urlapp "github.com/labib0x9/short/internal/app/url"
	"github.com/labib0x9/short/internal/infra/postgres"
	"github.com/labib0x9/short/internal/infra/rabbitmq"
	"github.com/labib0x9/short/internal/infra/redis"
	redis_cache "github.com/labib0x9/short/internal/infra/redis/cache"
)

// func ptr(t time.Time) *time.Time { return &t }

func TestShorten(t *testing.T) {

	cnf := config.GetConfig("../../../../../.env")

	pgConn := postgres.NewPostgresConn(cnf.PostgreSQL)
	defer pgConn.Close()

	redisClient := redis.Setup(cnf.Redis)
	defer redisClient.Close()

	rabbitMq := rabbitmq.NewRabbitMQ(cnf.RabbitMq)
	defer rabbitMq.Close()

	txMngr := postgres.NewTxManager(pgConn)
	urlRepo := postgres.NewUrlRepository(pgConn)
	analysisRepo := postgres.NewAnalysisRepository(pgConn)
	cacheRepo := redis_cache.NewCache(redisClient)

	srv := urlapp.NewService(urlRepo, analysisRepo, txMngr, cacheRepo, rabbitMq, cnf)
	validate := validator.New()
	handler := NewHandler(srv, validate)

	cases := []struct {
		req    urlRequest
		code   int
		serial string
	}{
		{serial: "1", req: urlRequest{Url: "https://example.com/page"}, code: 201},
		{serial: "2", req: urlRequest{Url: "http://localhost:3000/short", ExpireAt: new(time.Now().Add(10 * time.Minute))}, code: 201},
		{serial: "3", req: urlRequest{Url: "ftp://files.example.com/report.pdf"}, code: 201},
		{serial: "4", req: urlRequest{Url: "example.com"}, code: 400},
		{serial: "5", req: urlRequest{Url: "https://#"}, code: 400},
		{serial: "6", req: urlRequest{Url: "//example.com/page"}, code: 400},
		{serial: "7", req: urlRequest{Url: ""}, code: 400},
		{serial: "8", req: urlRequest{}, code: 400},
		{serial: "9", req: urlRequest{Url: "https://example.com", ExpireAt: new(time.Now().Add(-10 * time.Minute))}, code: 400},
		{serial: "10", req: urlRequest{Url: "https://example.com", ExpireAt: new(time.Now())}, code: 400},
	}

	for _, tc := range cases {
		t.Run(tc.serial, func(t *testing.T) {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(tc.req); err != nil {
				t.Error("Unexpected json encoding errorr", err)
			}

			r := httptest.NewRequest(http.MethodPost, "/short", &buf)
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.Shorten(w, r)

			resp := w.Result()
			if resp.StatusCode != tc.code {
				t.Errorf("Serial: %s, Wanted: %d, Got: %d", tc.serial, tc.code, resp.StatusCode)
			}
		})
	}
}
