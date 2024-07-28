package adapters

import (
	"context"
	"sync"
	"tickets/models"
)

type ReceiptsClientMock struct {
	mock sync.Mutex

	IssuedReceipts []models.IssueReceiptRequest
}

func (rcm *ReceiptsClientMock) IssueReceipt(_ context.Context, request models.IssueReceiptRequest) error {
	rcm.mock.Lock()
	defer rcm.mock.Unlock()

	rcm.IssuedReceipts = append(rcm.IssuedReceipts, request)
	return nil
}
