package event

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

func (h Handler) IssueReceipt(ctx context.Context, event *entities.TicketBookingConfirmed_v1) error {
	log.FromContext(ctx).Info("Issuing receipt")

	request := entities.IssueReceiptRequest{
		TicketID:       event.TicketID,
		Price:          event.Price,
		IdempotencyKey: event.Header.IdempotencyKey,
	}

	resp, err := h.receiptsService.IssueReceipt(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to issue receipt: %w", err)
	}

	return h.eventBus.Publish(ctx, entities.TicketReceiptIssued_v1{
		Header:        entities.NewEventHeaderWithIdempotencyKey(event.Header.IdempotencyKey),
		TicketID:      event.TicketID,
		ReceiptNumber: resp.ReceiptNumber,
		IssuedAt:      resp.IssuedAt,
	})
}
