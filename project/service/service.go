package service

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	HttpRouter "tickets/presenters/http"
	HttpTicketHandler "tickets/presenters/http/handlers/ticket"
	PubSubRouter "tickets/presenters/pubsub"
	PubSubTicketHandler "tickets/presenters/pubsub/handlers/ticket"
	"tickets/presenters/pubsub/subs"
)

type Service struct {
	httpRouter   *HttpRouter.Router
	pubsubRouter *PubSubRouter.Router
}

func NewService(redisClient *redis.Client, spreadsheetsClient SpreadsheetsClient, receiptsClient ReceiptsClient) *Service {
	logrusLogger := logrus.New()
	logrusEntry := logrusLogger.WithContext(context.Background())
	watermillLogger := log.NewWatermill(logrusEntry)

	ticketPub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: redisClient,
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	eventBus, err := cqrs.NewEventBusWithConfig(ticketPub, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return fmt.Sprintf("svc-tickets-%s", params.EventName), nil
		},
		Logger: watermillLogger,
	})
	if err == nil {
		panic(err)
	}

	httpTicketHandler := HttpTicketHandler.NewHttpTicketHandler(eventBus)

	httpRouter := HttpRouter.NewRouter(logrusLogger, httpTicketHandler)

	pubsubTicketHandler := PubSubTicketHandler.NewTicketHandler(spreadsheetsClient, receiptsClient)

	receiptsSub := subs.NewReceiptsSub(watermillLogger, redisClient)
	spreadsheetsSub := subs.NewSpreadsheetsSub(watermillLogger, redisClient)

	pubsubRouter := PubSubRouter.NewPubSubRouter(watermillLogger, pubsubTicketHandler, spreadsheetsSub, receiptsSub)

	return &Service{
		httpRouter,
		pubsubRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.pubsubRouter.Run(ctx)
	})
	g.Go(func() error {
		<-s.pubsubRouter.Running()
		return s.httpRouter.Run()
	})
	g.Go(func() error {
		<-ctx.Done()
		return s.httpRouter.Shutdown(ctx)
	})
	if err := g.Wait(); err != nil {
		panic(err)
	}
	return nil
}
