package command

import (
	"context"
	"tickets/entities"
)

func (h Handler) TicketRefund(ctx context.Context, ticket *entities.RefundTicket) error {
	reason := "customer requested refund"
	if err := h.refundsService.RefundReceipt(ctx, entities.VoidReceipt{
		TicketID:       ticket.TicketID,
		Reason:         reason,
		IdempotencyKey: ticket.Header.IdempotencyKey,
	}); err != nil {
		return err
	}

	if err := h.refundsService.VoidReceipt(ctx, entities.VoidReceipt{
		TicketID:       ticket.TicketID,
		Reason:         reason,
		IdempotencyKey: ticket.Header.IdempotencyKey,
	}); err != nil {
		return err
	}

	return nil
}
