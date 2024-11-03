package event

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"tickets/entities"
	"time"
)

func (h Handler) StoreTicket(ctx context.Context, event *entities.TicketBookingConfirmed_v1) error {
	log.FromContext(ctx).Info("Storing ticket")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return h.ticketRepo.Add(ctx, entities.Ticket{
		TicketID: event.TicketID,
		Price: entities.Money{
			Amount:   event.Price.Amount,
			Currency: event.Price.Currency,
		},
		CustomerEmail: event.CustomerEmail,
	})
}
