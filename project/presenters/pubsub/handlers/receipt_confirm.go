package handlers

import (
	"context"
	"tickets/models"
)

type ReceiptConfirmedHandler struct {
	receiptsClient ReceiptsClient
}

func NewReceiptConfirmedHandler(receiptsClient ReceiptsClient) *ReceiptConfirmedHandler {
	return &ReceiptConfirmedHandler{receiptsClient}
}

func (r *ReceiptConfirmedHandler) HandlerName() string {
	return "ReceiptConfirmedHandler"
}

func (r *ReceiptConfirmedHandler) NewEvent() interface{} {
	return &models.TicketBookingConfirmed{}
}

func (r *ReceiptConfirmedHandler) Handle(ctx context.Context, event any) error {
	e := event.(*models.TicketBookingConfirmed)
	if e.Price.Currency == "" {
		e.Price.Currency = "USD"
	}
	if err := r.receiptsClient.IssueReceipt(
		ctx,
		models.IssueReceiptRequest{
			TicketID: e.TicketID,
			Price:    e.Price,
		}); err != nil {
		return err
	}
	return nil
}
