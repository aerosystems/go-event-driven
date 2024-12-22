package command

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"tickets/db"
	"tickets/entities"
)

func (h Handler) BookShowTickets(ctx context.Context, command *entities.BookShowTickets) error {
	err := h.bookingsRepo.AddBooking(ctx, entities.Booking{
		BookingID:       command.BookingID,
		ShowID:          command.ShowId,
		NumberOfTickets: command.NumberOfTickets,
		CustomerEmail:   command.CustomerEmail,
	})
	if errors.Is(err, db.ErrBookingAlreadyExists) {
		// now AddBooking is called via Pub/Sub, we are taking into account at-least-once delivery
		return nil
	}
	// we want to publish BookingFailed_v1 only if there are no places left
	// in other scenario we assume that it's a temporary error and we want to retry
	// if it's not a temporary error, our alerting system will notify us about spinning message
	if errors.Is(err, db.ErrNoPlacesLeft) {
		// BookingMade_v1 is published by the bookingsRepo (via outbox)
		// we need to just publish BookingFailed_v1
		// no outbox is required here (as nothing is written to the database)
		publishErr := h.eventBus.Publish(ctx, entities.BookingFailed_v1{
			Header:        entities.NewEventHeader(),
			BookingID:     command.BookingID,
			FailureReason: err.Error(),
		})
		if publishErr != nil {
			return fmt.Errorf("failed to publish BookingFailed_v1 event: %w", publishErr)
		}
	}

	return err
}
