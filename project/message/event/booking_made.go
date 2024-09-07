package event

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/google/uuid"
	"tickets/entities"
)

func (h Handler) BookingMade(ctx context.Context, event *entities.BookingMade) error {
	log.FromContext(ctx).Info("Booking made")

	show, err := h.showRepo.Get(ctx, event.ShowId.String())
	if err != nil {
		return fmt.Errorf("could not get show: %w", err)
	}
	booking := entities.DeadNationBooking{
		BookingID:         event.BookingID,
		DeadNationEventID: uuid.MustParse(show.DeadNationID),
		CustomerEmail:     event.CustomerEmail,
		NumberOfTickets:   event.NumberOfTickets,
	}
	return h.deadNationService.Notify(ctx, booking)
}
