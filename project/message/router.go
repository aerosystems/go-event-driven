package message

import (
	"tickets/message/event"
	"tickets/message/outbox"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewWatermillRouter(
	postgresSubscriber message.Subscriber,
	publisher message.Publisher,
	eventProcessorConfig cqrs.EventProcessorConfig,
	eventHandler event.Handler,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	useMiddlewares(router, watermillLogger)

	outbox.AddForwarderHandler(postgresSubscriber, publisher, router, watermillLogger)

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, eventProcessorConfig)
	if err != nil {
		panic(err)
	}

	if err := eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"AppendToTracker",
			eventHandler.AppendToTracker,
		),
		cqrs.NewEventHandler(
			"TicketRefundToSheet",
			eventHandler.TicketRefundToSheet,
		),
		cqrs.NewEventHandler(
			"IssueReceipt",
			eventHandler.IssueReceipt,
		),
		cqrs.NewEventHandler(
			"StoreTicket",
			eventHandler.StoreTicket,
		),
		cqrs.NewEventHandler(
			"RemoveCanceledTicket",
			eventHandler.RemoveCanceledTicket,
		),
		cqrs.NewEventHandler(
			"PrintTicket",
			eventHandler.PrintTicket,
		),
		cqrs.NewEventHandler(
			"BookingMade",
			eventHandler.BookingMade,
		),
	); err != nil {
		panic(err)
	}

	return router
}
