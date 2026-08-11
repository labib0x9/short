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

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("failed to load location %s: %v", name, err)
	}
	return loc
}

func TestGet(t *testing.T) {
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

	liveResult, err := srv.Shorten(t.Context(), "https://example.com/yoyoyo", nil, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:153.0) Gecko/20100101 Firefox/153.0")
	if err != nil {
		t.Fatalf("fixture setup: Shorten() failed: %v", err)
	}

	nyLoc := mustLoadLocation(t, "America/New_York")

	expiredShort := "exp-" + uuid.New().String()[:8]
	if err := urlRepo.Create(t.Context(), urldomain.Url{
		URL:      "https://example.com/lalallalala",
		ShortURL: expiredShort,
		ExpireAt: new(time.Now().Add(-1 * time.Hour)),
	}); err != nil {
		t.Fatalf("fixture setup: expired url insert failed: %v", err)
	}

	notExpiredNY := time.Now().In(nyLoc).Add(2 * time.Hour)
	notExpiredShort := "exp-" + uuid.New().String()[:8]
	if err := urlRepo.Create(t.Context(), urldomain.Url{
		URL:      "https://example.com/lalallalala",
		ShortURL: notExpiredShort,
		ExpireAt: &notExpiredNY,
	}); err != nil {
		t.Fatalf("fixture setup: not-expired (NY tz) url insert failed: %v", err)
	}

	expiredNY := time.Now().In(nyLoc).Add(-2 * time.Hour)
	expiredShortDiff := "exp-" + uuid.New().String()[:8]
	if err := urlRepo.Create(t.Context(), urldomain.Url{
		URL:      "https://example.com/lalallalala",
		ShortURL: expiredShortDiff,
		ExpireAt: &expiredNY,
	}); err != nil {
		t.Fatalf("fixture setup: expired (NY tz) url insert failed: %v", err)
	}

	cases := []struct {
		serial    string
		shortCode string
		setPath   bool
		wantCode  int
	}{
		{serial: "1", shortCode: liveResult.Code, setPath: true, wantCode: http.StatusFound},
		{serial: "2 (cache hit, same code again)", shortCode: liveResult.Code, setPath: true, wantCode: http.StatusFound},
		{serial: "3", shortCode: expiredShort, setPath: true, wantCode: http.StatusGone},
		{serial: "4", shortCode: "doesnotexist", setPath: true, wantCode: http.StatusNotFound},
		{serial: "5", shortCode: "", setPath: false, wantCode: http.StatusBadRequest},
		{serial: "6 (expired, constructed in America/New_York tz)", shortCode: expiredShortDiff, setPath: true, wantCode: http.StatusGone},
		{serial: "7 (not expired, constructed in America/New_York tz)", shortCode: notExpiredShort, setPath: true, wantCode: http.StatusFound},
	}

	for _, tc := range cases {
		t.Run(tc.serial, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/"+tc.shortCode, nil)
			if tc.setPath {
				r.SetPathValue("code", tc.shortCode)
			}

			w := httptest.NewRecorder()
			handler.Get(w, r)

			resp := w.Result()
			if resp.StatusCode != tc.wantCode {
				t.Errorf("Serial: %s, Wanted: %d, Got: %d", tc.serial, tc.wantCode, resp.StatusCode)
			}
		})
	}
}
