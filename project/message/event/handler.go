package event

import (
	"context"
	"tickets/entities"
)

type Handler struct {
	spreadsheetsService SpreadsheetsAPI
	receiptsService     ReceiptsService
	filesService        FilesService
	ticketRepo          TicketRepository
}

func NewHandler(
	spreadsheetsService SpreadsheetsAPI,
	receiptsService ReceiptsService,
	filesService FilesService,
	ticketRepo TicketRepository,
) Handler {
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

	return Handler{
		spreadsheetsService: spreadsheetsService,
		receiptsService:     receiptsService,
		ticketRepo:          ticketRepo,
		filesService:        filesService,
	}
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) (entities.IssueReceiptResponse, error)
}

type FilesService interface {
	PrintTicket(ctx context.Context, ticket entities.Ticket) error
}

type TicketRepository interface {
	Add(ctx context.Context, t entities.Ticket) error
	Remove(ctx context.Context, ticketID string) error
}
