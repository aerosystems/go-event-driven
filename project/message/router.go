package message

import (
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"tickets/db"
	"tickets/entities"
	"tickets/message/command"
	"tickets/message/event"
	"tickets/message/outbox"
)

func NewWatermillRouter(
	postgresSubscriber message.Subscriber,
	redisPublisher message.Publisher,
	redisSubscriber message.Subscriber,
	eventProcessorConfig cqrs.EventProcessorConfig,
	eventHandler event.Handler,
	commandProcessorConfig cqrs.CommandProcessorConfig,
	commandsHandler command.Handler,
	opsReadModel db.OpsBookingReadModel,
	dataLake db.DataLake,
	vipBundleProcessManager *entities.VipBundleProcessManager,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	useMiddlewares(router, watermillLogger)

	outbox.AddForwarderHandler(postgresSubscriber, redisPublisher, router, watermillLogger)

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, eventProcessorConfig)
	if err != nil {
		panic(err)
	}

	eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"BookPlaceInDeadNation",
			eventHandler.BookPlaceInDeadNation,
		),
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
			"PrintTicketHandler",
			eventHandler.PrintTicket,
		),
		cqrs.NewEventHandler(
			"StoreTickets",
			eventHandler.StoreTickets,
		),
		cqrs.NewEventHandler(
			"RemoveCanceledTicket",
			eventHandler.RemoveCanceledTicket,
		),
		cqrs.NewEventHandler(
			"ops_read_model.OnBookingMade",
			opsReadModel.OnBookingMade,
		),
		cqrs.NewEventHandler(
			"ops_read_model.IssueReceiptHandler",
			opsReadModel.OnTicketReceiptIssued,
		),
		cqrs.NewEventHandler(
			"ops_read_model.OnTicketBookingConfirmed",
			opsReadModel.OnTicketBookingConfirmed,
		),
		cqrs.NewEventHandler(
			"ops_read_model.OnTicketPrinted",
			opsReadModel.OnTicketPrinted,
		),
		cqrs.NewEventHandler(
			"ops_read_model.OnTicketRefunded",
			opsReadModel.OnTicketRefunded,
		),
		cqrs.NewEventHandler(
			"vip_bundle_process_manager.OnVipBundleInitialized",
			vipBundleProcessManager.OnVipBundleInitialized,
		),
		cqrs.NewEventHandler(
			"vip_bundle_process_manager.OnBookingMade",
			vipBundleProcessManager.OnBookingMade,
		),
		cqrs.NewEventHandler(
			"vip_bundle_process_manager.OnBookingFailed",
			vipBundleProcessManager.OnBookingFailed,
		),
		cqrs.NewEventHandler(
			"vip_bundle_process_manager.OnFlightBooked",
			vipBundleProcessManager.OnFlightBooked,
		),
		cqrs.NewEventHandler(
			"vip_bundle_process_manager.OnFlightBookingFailed",
			vipBundleProcessManager.OnFlightBookingFailed,
		),
		cqrs.NewEventHandler(
			"vip_bundle_process_manager.OnTaxiBooked",
			vipBundleProcessManager.OnTaxiBooked,
		),
		cqrs.NewEventHandler(
			"vip_bundle_process_manager.OnTaxiBookingFailed",
			vipBundleProcessManager.OnTaxiBookingFailed,
		),
	)

	commandProcessor, err := cqrs.NewCommandProcessorWithConfig(
		router,
		commandProcessorConfig,
	)
	if err != nil {
		panic(err)
	}

	commandProcessor.AddHandlers(
		cqrs.NewCommandHandler(
			"TicketRefund",
			commandsHandler.RefundTicket,
		),
		cqrs.NewCommandHandler(
			"BookShowTickets",
			commandsHandler.BookShowTickets,
		),
		cqrs.NewCommandHandler(
			"BookFlight",
			commandsHandler.BookFlight,
		),
		cqrs.NewCommandHandler(
			"BookTaxi",
			commandsHandler.BookTaxi,
		),
	)

	router.AddNoPublisherHandler(
		"events_splitter",
		"events",
		redisSubscriber,
		func(msg *message.Message) error {
			eventName := eventProcessorConfig.Marshaler.NameFromMessage(msg)
			if eventName == "" {
				return fmt.Errorf("cannot get event name from message")
			}

			return redisPublisher.Publish("events."+eventName, msg)
		},
	)

	router.AddNoPublisherHandler(
		"store_to_data_lake",
		"events",
		redisSubscriber,
		func(msg *message.Message) error {
			eventName := eventProcessorConfig.Marshaler.NameFromMessage(msg)
			if eventName == "" {
				return fmt.Errorf("cannot get event name from message")
			}

			// we just need to unmarshal event header, rest is stored as is
			type Event struct {
				Header entities.EventHeader `json:"header"`
			}

			var event Event
			if err := eventProcessorConfig.Marshaler.Unmarshal(msg, &event); err != nil {
				return fmt.Errorf("cannot unmarshal event: %w", err)
			}

			return dataLake.StoreEvent(
				msg.Context(),
				entities.DataLakeEvent{
					EventID:      event.Header.ID,
					PublishedAt:  event.Header.PublishedAt,
					EventName:    eventName,
					EventPayload: msg.Payload,
				},
			)
		},
	)

	return router
}
