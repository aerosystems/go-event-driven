package event

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"tickets/entities"
)

func (h Handler) RemoveCanceledTicket(ctx context.Context, event *entities.TicketBookingCanceled) error {
	log.FromContext(ctx).Info("Removing canceled ticket")

	return h.ticketRepo.Remove(ctx, event.TicketID)
}
