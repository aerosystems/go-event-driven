package event

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"tickets/entities"
)

func (h Handler) PrintTicket(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Info("Printing ticket")

	return h.filesService.PrintTicket(
		ctx,
		entities.Ticket{
			TicketID: event.TicketID,
			Price: entities.Money{
				Amount:   event.Price.Amount,
				Currency: event.Price.Currency,
			},
			CustomerEmail: event.CustomerEmail,
		})
}
