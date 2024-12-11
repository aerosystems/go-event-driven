package event

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"tickets/entities"
)

func (h Handler) RemoveCanceledTicket(ctx context.Context, event *entities.TicketBookingCanceled_v1) error {
	log.FromContext(ctx).Info("Storing ticket")

	return h.ticketsRepository.Remove(ctx, event.TicketID)
}
