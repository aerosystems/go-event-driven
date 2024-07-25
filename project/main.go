package main

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os/signal"
	"tickets/config"
	"tickets/infra/adapters"
	HttpRouter "tickets/presenters/http"
	HttpTicketHandler "tickets/presenters/http/handlers/ticket"
	"tickets/presenters/pubsub"
	PubSubTicketHandler "tickets/presenters/pubsub/handlers/ticket"
	"tickets/presenters/pubsub/subs"
)

func init() {
	log.Init(logrus.InfoLevel)
}

func main() {
	logger := watermill.NewStdLogger(false, false)
	cfg := config.NewConfig()

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

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
	})

	ticketPub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, log.NewWatermill(logrus.NewEntry(logrus.StandardLogger())))
	if err != nil {
		panic(err)
	}

	spreadsheetsClient := adapters.NewSpreadsheetsClient(apiClients)
	receiptsClient := adapters.NewReceiptsClient(apiClients)

	httpTicketHandler := HttpTicketHandler.NewHttpTicketHandler(ticketPub)

	httpRouter := HttpRouter.NewRouter(cfg.WebPort, httpTicketHandler)

	pubsubTicketHandler := PubSubTicketHandler.NewTicketHandler(spreadsheetsClient, receiptsClient)

	receiptsSub := subs.NewReceiptsSub(logger, rdb)
	spreadsheetsSub := subs.NewSpreadsheetsSub(logger, rdb)

	pubsubRouter := pubsub.NewPubSubRouter(logger, pubsubTicketHandler, spreadsheetsSub, receiptsSub)

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return pubsubRouter.Run()
	})
	g.Go(func() error {
		<-pubsubRouter.Running()
		return httpRouter.Run()
	})
	g.Go(func() error {
		<-ctx.Done()
		return httpRouter.Shutdown(ctx)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
