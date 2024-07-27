package tests_test

import (
	"context"
	"github.com/redis/go-redis/v9"
	"net/http"
	"testing"
	"tickets/config"
	"tickets/infra/adapters"
	"tickets/service"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	cfg := config.NewConfig()
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
	})
	spreadsheetsClientMock := &adapters.SpreadsheetsClientMock{}
	receiptsClientMock := &adapters.ReceiptsClientMock{}

	go func() {
		svc := service.NewService(redisClient, spreadsheetsClientMock, receiptsClientMock)
		assert.NoError(t, svc.Run(context.Background()))
	}()

	waitForHttpServer(t)
}

func waitForHttpServer(t *testing.T) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			resp, err := http.Get("http://localhost:8080/health")
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			if assert.Less(t, resp.StatusCode, 300, "API not ready, http status: %d", resp.StatusCode) {
				return
			}
		},
		time.Second*10,
		time.Millisecond*50,
	)
}
