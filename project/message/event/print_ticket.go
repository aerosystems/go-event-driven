package event

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"tickets/entities"
)

func (h Handler) PrintTicket(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Info("Printing ticket")

	fileID, err := h.filesService.PrintTicket(
		ctx,
		entities.Ticket{
			TicketID: event.TicketID,
			Price: entities.Money{
				Amount:   event.Price.Amount,
				Currency: event.Price.Currency,
			},
			CustomerEmail: event.CustomerEmail,
		})
	if err != nil {
		return fmt.Errorf("failed to print ticket: %w", err)
	}
	if fileID == "" {
		return nil
	}

	return h.eventBus.Publish(ctx, entities.TicketPrinted{
		Header:   entities.NewEventHeader(),
		TicketID: event.TicketID,
		FileName: fileID,
	})
}
