package main

import (
	"context"
	"github.com/google/uuid"
	"sync"
	"time"
)

type IssueReceiptRequest struct {
	TicketID string `json:"ticket_id"`
	Price    Money  `json:"price"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type IssueReceiptResponse struct {
	ReceiptNumber string    `json:"number"`
	IssuedAt      time.Time `json:"issued_at"`
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request IssueReceiptRequest) (IssueReceiptResponse, error)
}

type ReceiptsServiceMock struct {
	mock sync.Mutex

	IssuedReceipts []IssueReceiptRequest
}

func (rsm *ReceiptsServiceMock) IssueReceipt(_ context.Context, request IssueReceiptRequest) (IssueReceiptResponse, error) {
	rsm.mock.Lock()
	defer rsm.mock.Unlock()
	receiptNumber, err := uuid.NewUUID()
	if err != nil {
		return IssueReceiptResponse{}, nil
	}

	rsm.IssuedReceipts = append(rsm.IssuedReceipts, request)
	return IssueReceiptResponse{
		receiptNumber.String(),
		time.Now(),
	}, nil
}
