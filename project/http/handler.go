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
	showRepo              ShowRepository
	bookingRepo           BookingRepository
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, spreadsheetName string, row []string) error
}

type TicketRepository interface {
	GetAll(ctx context.Context) ([]entities.Ticket, error)
}

type ShowRepository interface {
	Create(ctx context.Context, show entities.Show) (string, error)
}

type BookingRepository interface {
	Create(ctx context.Context, booking entities.Booking) (string, error)
}
