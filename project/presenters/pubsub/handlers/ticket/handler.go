package PubSubTicketHandler

const brokenMsgUuid = "2beaf5bc-d5e4-4653-b075-2b36bbf28949"

type Handler struct {
	spreadsheetsClient SpreadsheetsClient
	receiptsClient     ReceiptsClient
}

func NewTicketHandler(spreadsheetsClient SpreadsheetsClient, receiptsClient ReceiptsClient) *Handler {
	return &Handler{
		spreadsheetsClient,
		receiptsClient,
	}
}
