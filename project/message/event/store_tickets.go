package event

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"tickets/entities"
)

func (h Handler) StoreTickets(ctx context.Context, event *entities.TicketBookingConfirmed_v1) error {
	log.FromContext(ctx).Info("Storing ticket")

	return h.ticketsRepository.Add(ctx, entities.Ticket{
		TicketID:      event.TicketID,
		Price:         event.Price,
		CustomerEmail: event.CustomerEmail,
	})
}
