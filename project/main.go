package main

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"net/http"
	"tickets/config"
	"tickets/infra/adapters"
	"tickets/service"
)

func init() {
	log.Init(logrus.InfoLevel)
}

func main() {
	cfg := config.NewConfig()

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
	})
	defer rdb.Close()

	apiClients, err := clients.NewClients(
		cfg.GatewayAddress,
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	spreadsheetsClient := adapters.NewSpreadsheetsClient(apiClients)
	receiptsClient := adapters.NewReceiptsClient(apiClients)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.NewService(rdb, spreadsheetsClient, receiptsClient)
	svc.Run(ctx)
}
