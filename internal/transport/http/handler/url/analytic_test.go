package url

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labib0x9/short/config"
	urlapp "github.com/labib0x9/short/internal/app/url"
	urldomain "github.com/labib0x9/short/internal/domain/url"
	"github.com/labib0x9/short/internal/infra/postgres"
	"github.com/labib0x9/short/internal/infra/rabbitmq"
	"github.com/labib0x9/short/internal/infra/redis"
	redis_cache "github.com/labib0x9/short/internal/infra/redis/cache"
)

func TestAnalysis(t *testing.T) {
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
	handler := NewHandler(srv, validator.New())

	clicked, err := srv.Shorten(t.Context(), "https://example.com/analysis-clicked", nil, "test-agent")
	if err != nil {
		t.Fatalf("fixture setup: Shorten() failed: %v", err)
	}
	clickedFull, err := urlRepo.GetByShortCode(t.Context(), clicked.Code)
	if err != nil {
		t.Fatalf("fixture setup: GetByShortCode() failed: %v", err)
	}
	if err := analysisRepo.Create(t.Context(), urldomain.Click{
		UrlId:      clickedFull.Id,
		Referer:    "https://google.com",
		Country:    "unknown",
		DeviceType: "Desktop",
		Os:         "Linux",
		Browser:    "Firefox",
		ClickedAt:  time.Now(),
	}); err != nil {
		t.Fatalf("fixture setup: click insert failed: %v", err)
	}

	unclicked, err := srv.Shorten(t.Context(), "https://example.com/analysis-unclicked", nil, "test-agent")
	if err != nil {
		t.Fatalf("fixture setup: Shorten() failed: %v", err)
	}

	expiredShort := "ana-" + uuid.New().String()[:8]
	pastExpiry := time.Now().Add(-1 * time.Hour)
	if err := urlRepo.Create(t.Context(), urldomain.Url{
		URL:       "https://example.com/analysis-expired",
		ShortURL:  expiredShort,
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpireAt:  &pastExpiry,
	}); err != nil {
		t.Fatalf("fixture setup: expired url insert failed: %v", err)
	}

	cases := []struct {
		serial   string
		code     string
		setPath  bool
		wantCode int
	}{
		{serial: "1", code: clicked.Code, setPath: true, wantCode: http.StatusOK},
		{serial: "2 (zero clicks)", code: unclicked.Code, setPath: true, wantCode: http.StatusOK},
		{serial: "3", code: expiredShort, setPath: true, wantCode: http.StatusGone},
		{serial: "4", code: "doesnotexist", setPath: true, wantCode: http.StatusNotFound},
		{serial: "5", code: "", setPath: false, wantCode: http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.serial, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/"+tc.code+"/stat", nil)
			if tc.setPath {
				r.SetPathValue("code", tc.code)
			}

			w := httptest.NewRecorder()
			handler.Analysis(w, r)

			resp := w.Result()
			if resp.StatusCode != tc.wantCode {
				t.Errorf("Serial: %s, Wanted: %d, Got: %d", tc.serial, tc.wantCode, resp.StatusCode)
			}
		})
	}
}
