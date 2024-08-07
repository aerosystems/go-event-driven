package event

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"tickets/entities"
)

func (h Handler) PrintTicket(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Info("Printing ticket")

	h.eventBus.Publish(ctx, entities.TicketPrinted{
		Header:   event.Header,
		TicketID: event.TicketID,
		FileName: fmt.Sprintf("%s-ticket.html", event.TicketID),
	})

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
