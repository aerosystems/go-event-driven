package http

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/google/uuid"
	"tickets/entities"
)

type Handler struct {
	eventBus              *cqrs.EventBus
	commandBus            *cqrs.CommandBus
	spreadsheetsAPIClient SpreadsheetsAPI
	ticketsRepo           TicketsRepository

	opsBookingReadModel  OpsBookingReadModel
	showsRepository      ShowsRepository
	bookingsRepository   BookingsRepository
	vipBundlesRepository VipBundlesRepository
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, spreadsheetName string, row []string) error
}

type TicketsRepository interface {
	FindAll(ctx context.Context) ([]entities.Ticket, error)
}

type ShowsRepository interface {
	AddShow(ctx context.Context, show entities.Show) error
	AllShows(ctx context.Context) ([]entities.Show, error)
	ShowByID(ctx context.Context, showID uuid.UUID) (entities.Show, error)
}

type OpsBookingReadModel interface {
	AllReservations(receiptIssueDateFilter string) ([]entities.OpsBooking, error)
	ReservationReadModel(ctx context.Context, id string) (entities.OpsBooking, error)
}
