package command

import (
	"fmt"
	"golang.org/x/net/context"
	"tickets/entities"
)

func (h Handler) BookTaxi(ctx context.Context, command *entities.BookTaxi) error {
	resp, err := h.transportationServiceClient.BookTaxi(ctx, entities.BookTaxiRequest{
		CustomerEmail:      command.CustomerEmail,
		NumberOfPassengers: command.NumberOfPassengers,
		PassengerName:      command.CustomerName,
		ReferenceId:        command.ReferenceID,
		IdempotencyKey:     command.IdempotencyKey,
	})
	if err != nil {
		return fmt.Errorf("failed to book taxi: %w", err)
	}

	err = h.eventBus.Publish(ctx, entities.TaxiBooked_v1{
		Header:        entities.NewEventHeader(),
		TaxiBookingID: resp.TaxiBookingId,
		ReferenceID:   command.ReferenceID,
	})
	if err != nil {
		return fmt.Errorf("failed to publish TaxiBooked_v1 event: %w", err)
	}

	return nil
}
