package command

import (
	"fmt"
	"golang.org/x/net/context"
	"tickets/entities"
)

func (h Handler) BookFlight(ctx context.Context, command *entities.BookFlight) error {
	resp, err := h.transportationServiceClient.BookFlight(ctx, entities.BookFlightTicketRequest{
		CustomerEmail:  command.CustomerEmail,
		FlightID:       command.FlightID,
		PassengerNames: command.Passengers,
		ReferenceId:    command.ReferenceID,
		IdempotencyKey: command.IdempotencyKey,
	})
	if err != nil {
		return fmt.Errorf("failed to void receipt: %w", err)
	}

	err = h.eventBus.Publish(ctx, entities.FlightBooked_v1{
		Header:      entities.NewEventHeader(),
		FlightID:    command.FlightID,
		TicketIDs:   resp.TicketIds,
		ReferenceID: command.ReferenceID,
	})
	if err != nil {
		return fmt.Errorf("failed to publish FlightBooked_v1 event: %w", err)
	}

	return nil
}
