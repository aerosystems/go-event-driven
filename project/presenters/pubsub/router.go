package PubSubRouter

import (
	"context"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	PubSubTicketHandler "tickets/presenters/pubsub/handlers/ticket"
)

type Router struct {
	router          *message.Router
	logger          watermill.LoggerAdapter
	ticketHandler   *PubSubTicketHandler.Handler
	spreadsheetsSub *redisstream.Subscriber
	receiptsSub     *redisstream.Subscriber
}

func NewPubSubRouter(logger watermill.LoggerAdapter, ticketHandler *PubSubTicketHandler.Handler, spreadsheetsSub *redisstream.Subscriber, receiptsSub *redisstream.Subscriber) *Router {
	var err error
	var r Router

	r = Router{
		logger:          logger,
		ticketHandler:   ticketHandler,
		spreadsheetsSub: spreadsheetsSub,
		receiptsSub:     receiptsSub,
	}

	r.router, err = message.NewRouter(message.RouterConfig{}, r.logger)
	if err != nil {
		panic(err)
	}

	r.router.AddMiddleware(TracingMiddleware, LoggingMiddleware, HandleErrorsMiddleware, ExponentialBackoffMiddleware)

	r.router.AddNoPublisherHandler(
		"receipt-handler",
		"TicketBookingConfirmed",
		r.receiptsSub,
		r.ticketHandler.ReceiptConfirm)

	r.router.AddNoPublisherHandler(
		"spreadsheet-confirmed-handler",
		"TicketBookingConfirmed",
		r.spreadsheetsSub,
		r.ticketHandler.SpreadsheetConfirm,
	)

	r.router.AddNoPublisherHandler(
		"spreadsheet-canceled-handler",
		"TicketBookingCanceled",
		r.spreadsheetsSub,
		r.ticketHandler.SpreadsheetCancel)

	return &r
}

func (r *Router) Run() error {
	return r.router.Run(context.Background())
}

func (r *Router) Running() chan struct{} {
	return r.router.Running()
}
