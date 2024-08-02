package handlers

import (
	"context"
	"tickets/models"
)

type SpreadsheetConfirmedHandler struct {
	spreadsheetsClient SpreadsheetsClient
}

func NewSpreadsheetConfirmedHandler(spreadsheetsClient SpreadsheetsClient) *SpreadsheetConfirmedHandler {
	return &SpreadsheetConfirmedHandler{
		spreadsheetsClient,
	}
}

func (h *SpreadsheetConfirmedHandler) HandlerName() string {
	return "SpreadsheetCanceledHandler"
}

func (h *SpreadsheetConfirmedHandler) NewEvent() interface{} {
	return &models.TicketBookingCanceled{}
}

func (h *SpreadsheetConfirmedHandler) Handle(ctx context.Context, event any) error {
	e := event.(*models.TicketBookingCanceled)
	if err := h.spreadsheetsClient.AppendRow(ctx, "tickets-to-print", []string{e.TicketID, e.CustomerEmail, e.Price.Amount, e.Price.Currency}); err != nil {
		return err
	}
	return nil
}
