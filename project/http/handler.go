package http

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/entities"
)

type Handler struct {
	eventBus              *cqrs.EventBus
	spreadsheetsAPIClient SpreadsheetsAPI
	ticketRepo            TicketRepository
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, spreadsheetName string, row []string) error
}

type TicketRepository interface {
	GetAll(ctx context.Context) ([]entities.Ticket, error)
}
