package handlers

import (
	"context"
	"tickets/models"
)

type SpreadsheetCanceledHandler struct {
	spreadsheetsClient SpreadsheetsClient
}

func NewSpreadsheetCanceledHandler(spreadsheetsClient SpreadsheetsClient) *SpreadsheetCanceledHandler {
	return &SpreadsheetCanceledHandler{
		spreadsheetsClient,
	}
}

func (h *SpreadsheetCanceledHandler) HandlerName() string {
	return "SpreadsheetCanceledHandler"
}

func (h *SpreadsheetCanceledHandler) NewEvent() interface{} {
	return &models.TicketBookingCanceled{}
}

func (h *SpreadsheetCanceledHandler) Handle(ctx context.Context, event any) error {
	e := event.(*models.TicketBookingCanceled)
	if err := h.spreadsheetsClient.AppendRow(ctx, "tickets-to-refund", []string{e.TicketID, e.CustomerEmail, e.Price.Amount, e.Price.Currency}); err != nil {
		return err
	}
	return nil
}
