package adapters

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/receipts"
	"net/http"
	"sync"
	"tickets/models"
)

type ReceiptsClient struct {
	clients *clients.Clients
}

func NewReceiptsClient(clients *clients.Clients) *ReceiptsClient {
	return &ReceiptsClient{
		clients: clients,
	}
}

func (c ReceiptsClient) IssueReceipt(ctx context.Context, request models.IssueReceiptRequest) error {
	body := receipts.PutReceiptsJSONRequestBody{
		TicketId: request.TicketID,
		Price: receipts.Money{
			MoneyAmount:   request.Price.Amount,
			MoneyCurrency: request.Price.Currency,
		},
	}

	receiptsResp, err := c.clients.Receipts.PutReceiptsWithResponse(ctx, body)
	if err != nil {
		return err
	}
	if receiptsResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", receiptsResp.StatusCode())
	}

	return nil
}

type ReceiptsClientMock struct {
	mock sync.Mutex

	IssueReceipts []models.IssueReceiptRequest
}

func (rcm *ReceiptsClientMock) IssueReceipt(_ context.Context, request models.IssueReceiptRequest) error {
	rcm.mock.Lock()
	defer rcm.mock.Unlock()

	rcm.IssueReceipts = append(rcm.IssueReceipts, request)
	return nil
}
