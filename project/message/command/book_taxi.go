package command

import (
	"errors"
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
	if errors.Is(err, entities.ErrNoTaxiAvailable) {
		err = h.eventBus.Publish(ctx, entities.TaxiBookingFailed_v1{
			Header:        entities.NewEventHeader(),
			FailureReason: err.Error(),
			ReferenceID:   command.ReferenceID,
		})
		if err != nil {
			return fmt.Errorf("failed to publish TaxiBookingFailed_v1 event: %w", err)
		}

		return nil
	}
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
