package handlers

import (
	"context"
	"fmt"
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
	return "SpreadsheetConfirmedHandler"
}

func (h *SpreadsheetConfirmedHandler) NewEvent() interface{} {
	return &models.TicketBookingConfirmed{}
}

func (h *SpreadsheetConfirmedHandler) Handle(ctx context.Context, event any) error {
	fmt.Println("!!! SpreadsheetConfirmedHandler")
	e := event.(*models.TicketBookingConfirmed)
	if err := h.spreadsheetsClient.AppendRow(ctx, "tickets-to-print", []string{e.TicketID, e.CustomerEmail, e.Price.Amount, e.Price.Currency}); err != nil {
		return err
	}
	return nil
}
