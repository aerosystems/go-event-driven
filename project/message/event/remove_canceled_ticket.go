package event

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

func (h Handler) RemoveCanceledTicket(ctx context.Context, ticketID string) error {
	log.FromContext(ctx).Info("Removing canceled ticket")

	return h.ticketRepo.Remove(ctx, ticketID)
}
