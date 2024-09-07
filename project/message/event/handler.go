package event

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/entities"
)

type Handler struct {
	eventBus            *cqrs.EventBus
	spreadsheetsService SpreadsheetsAPI
	receiptsService     ReceiptsService
	filesService        FilesService
	ticketRepo          TicketRepository
	showRepo            ShowRepository
	deadNationService   DeadNationService
}

func NewHandler(
	eventBus *cqrs.EventBus,
	spreadsheetsService SpreadsheetsAPI,
	receiptsService ReceiptsService,
	filesService FilesService,
	ticketRepo TicketRepository,
	showRepo ShowRepository,
	deadNationService DeadNationService,
) Handler {
	if eventBus == nil {
		panic("missing eventBus")
	}
	if spreadsheetsService == nil {
		panic("missing spreadsheetsService")
	}
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if filesService == nil {
		panic("missing filesService")
	}
	if ticketRepo == nil {
		panic("missing ticketRepo")
	}
	if showRepo == nil {
		panic("missing showRepo")
	}
	if deadNationService == nil {
		panic("missing deadNationService")
	}

	return Handler{
		eventBus:            eventBus,
		spreadsheetsService: spreadsheetsService,
		receiptsService:     receiptsService,
		ticketRepo:          ticketRepo,
		showRepo:            showRepo,
		filesService:        filesService,
		deadNationService:   deadNationService,
	}
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) (entities.IssueReceiptResponse, error)
}

type FilesService interface {
	PrintTicket(ctx context.Context, ticket entities.Ticket) (string, error)
}

type TicketRepository interface {
	Add(ctx context.Context, t entities.Ticket) error
	Remove(ctx context.Context, ticketID string) error
}

type DeadNationService interface {
	Notify(ctx context.Context, booking entities.DeadNationBooking) error
}

type ShowRepository interface {
	Get(ctx context.Context, showID string) (entities.Show, error)
}
