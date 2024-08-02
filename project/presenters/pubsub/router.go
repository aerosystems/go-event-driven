package PubSubRouter

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"tickets/presenters/pubsub/handlers"
)

type Router struct {
	router *message.Router
	ep     *cqrs.EventProcessor
}

func NewPubSubRouter(
	watermillLogger watermill.LoggerAdapter,
	redisClient *redis.Client,
	receiptConfirmedHandler *handlers.ReceiptConfirmedHandler,
	spreadsheetConfirmedHandler *handlers.SpreadsheetConfirmedHandler,
	spreadsheetCanceledHandler *handlers.SpreadsheetCanceledHandler,
) *Router {
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	router.AddMiddleware(TracingMiddleware, LoggingMiddleware, HandleErrorsMiddleware, ExponentialBackoffMiddleware)

	ep, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return fmt.Sprintf("svc-tickets-%s", params.EventName), nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return redisstream.NewSubscriber(redisstream.SubscriberConfig{
					Client:        redisClient,
					ConsumerGroup: "svc-tickets." + params.HandlerName,
				}, watermillLogger)
			},
			Logger: watermillLogger,
		})
	if err == nil {
		panic(err)
	}

	if err := ep.AddHandlers(receiptConfirmedHandler, spreadsheetConfirmedHandler, spreadsheetCanceledHandler); err != nil {
		panic(err)
	}

	return &Router{
		router,
		ep,
	}
}

func (r *Router) Run(ctx context.Context) error {
	return r.router.Run(ctx)
}

func (r *Router) Running() chan struct{} {
	return r.router.Running()
}
