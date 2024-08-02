package service

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"tickets/common"
	HttpRouter "tickets/presenters/http"
	HttpTicketHandler "tickets/presenters/http/handlers/ticket"
	PubSubRouter "tickets/presenters/pubsub"
	"tickets/presenters/pubsub/handlers"
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

	publisher := common.CorrelationPublisherDecorator{Publisher: ticketPub}

	eventBus, err := cqrs.NewEventBusWithConfig(publisher, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return params.EventName, nil
		},
		Marshaler: cqrs.JSONMarshaler{
			GenerateName: cqrs.StructName,
		},
		Logger: watermillLogger,
	})
	if err != nil {
		panic(err)
	}

	httpTicketHandler := HttpTicketHandler.NewHttpTicketHandler(eventBus)

	httpRouter := HttpRouter.NewRouter(logrusLogger, httpTicketHandler)

	receiptConfirmedHandler := handlers.NewReceiptConfirmedHandler(receiptsClient)
	spreadsheetConfirmedHandler := handlers.NewSpreadsheetConfirmedHandler(spreadsheetsClient)
	spreadsheetCanceledHandler := handlers.NewSpreadsheetCanceledHandler(spreadsheetsClient)

	pubsubRouter := PubSubRouter.NewPubSubRouter(watermillLogger, redisClient)

	if err := pubsubRouter.RegisterEventHandlers(receiptConfirmedHandler, spreadsheetConfirmedHandler, spreadsheetCanceledHandler); err != nil {
		panic(err)
	}

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
