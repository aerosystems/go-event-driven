package PubSubRouter

import (
	"context"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	PubSubTicketHandler "tickets/presenters/pubsub/handlers/ticket"
)

type Router struct {
	router *message.Router
}

func NewPubSubRouter(logger watermill.LoggerAdapter, ticketHandler *PubSubTicketHandler.Handler, spreadsheetsSub *redisstream.Subscriber, receiptsSub *redisstream.Subscriber) *Router {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	router.AddMiddleware(TracingMiddleware, LoggingMiddleware, HandleErrorsMiddleware, ExponentialBackoffMiddleware)

	router.AddNoPublisherHandler(
		"receipt-handler",
		"TicketBookingConfirmed",
		receiptsSub,
		ticketHandler.ReceiptConfirm)

	router.AddNoPublisherHandler(
		"spreadsheet-confirmed-handler",
		"TicketBookingConfirmed",
		spreadsheetsSub,
		ticketHandler.SpreadsheetConfirm,
	)

	router.AddNoPublisherHandler(
		"spreadsheet-canceled-handler",
		"TicketBookingCanceled",
		spreadsheetsSub,
		ticketHandler.SpreadsheetCancel)

	return &Router{
		router: router,
	}
}

func (r *Router) Run(ctx context.Context) error {
	return r.router.Run(ctx)
}

func (r *Router) Running() chan struct{} {
	return r.router.Running()
}
