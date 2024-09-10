package command

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"tickets/entities"
)

func (h Handler) TicketRefund(ctx context.Context, ticket *entities.Ticket) error {
	log.FromContext(ctx).Info("Refunding ticket")

	return nil
}
