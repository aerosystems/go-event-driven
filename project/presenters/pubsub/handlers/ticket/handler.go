package PubSubTicketHandler

import (
	"tickets/infra/adapters"
)

const brokenMsgUuid = "2beaf5bc-d5e4-4653-b075-2b36bbf28949"

type Handler struct {
	spreadsheetsClient *adapters.SpreadsheetsClient
	receiptsClient     *adapters.ReceiptsClient
}

func NewTicketHandler(spreadsheetsClient *adapters.SpreadsheetsClient, receiptsClient *adapters.ReceiptsClient) *Handler {
	return &Handler{
		spreadsheetsClient,
		receiptsClient,
	}
}
