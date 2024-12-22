package command

import (
	"errors"
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

	return err
}
