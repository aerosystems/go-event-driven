package http

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/entities"
)

type Handler struct {
	commandBus            *cqrs.CommandBus
	eventBus              *cqrs.EventBus
	spreadsheetsAPIClient SpreadsheetsAPI
	ticketRepo            TicketRepository
	showRepo              ShowRepository
	bookingRepo           BookingRepository
	opsReadModel          OpsReadModel
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
	AddBooking(ctx context.Context, booking entities.Booking) error
}

type OpsReadModel interface {
	AllReservations() ([]entities.OpsBooking, error)
	ReservationReadModel(ctx context.Context, bookingID string) (entities.OpsBooking, error)
}
