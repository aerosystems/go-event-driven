package api

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/payments"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/receipts"
	"net/http"
	"tickets/entities"
)

type RefundsServiceClient struct {
	// we are not mocking this client: it's pointless to use interface here
	clients *clients.Clients
}

func NewRefundsServiceClient(clients *clients.Clients) *ReceiptsServiceClient {
	if clients == nil {
		panic("NewReceiptsServiceClient: clients is nil")
	}

	return &ReceiptsServiceClient{clients: clients}
}

func (c ReceiptsServiceClient) RefundReceipt(ctx context.Context, request entities.VoidReceipt) error {
	res, err := c.clients.Payments.PutRefundsWithResponse(ctx, payments.PaymentRefundRequest{
		PaymentReference: request.TicketID,
		Reason:           request.Reason,
		DeduplicationId:  &request.IdempotencyKey,
	})
	if err != nil {
		return fmt.Errorf("failed to refund receipt: %w", err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to refund receipt: unexpected status code %d", res.StatusCode())
	}

	return nil
}

func (c ReceiptsServiceClient) VoidReceipt(ctx context.Context, request entities.VoidReceipt) error {
	res, err := c.clients.Receipts.PutVoidReceiptWithResponse(ctx, receipts.VoidReceiptRequest{
		Reason:       request.Reason,
		TicketId:     request.TicketID,
		IdempotentId: &request.IdempotencyKey,
	})
	if err != nil {
		return fmt.Errorf("failed to void receipt: %w", err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to void receipt: unexpected status code %d", res.StatusCode())
	}

	return nil
}
