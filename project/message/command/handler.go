package command

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/entities"
)

type Handler struct {
	commandBus      *cqrs.CommandBus
	paymentsService PaymentsService
	receiptService  ReceiptService
}

func NewHandler(commandBus *cqrs.CommandBus, paymentsService PaymentsService, receiptService ReceiptService) Handler {
	if commandBus == nil {
		panic("missing commandBus")
	}
	if paymentsService == nil {
		panic("missing paymentsService")
	}
	if receiptService == nil {
		panic("missing receiptService")
	}

	return Handler{
		commandBus:      commandBus,
		paymentsService: paymentsService,
		receiptService:  receiptService,
	}
}

type PaymentsService interface {
	RefundPayment(ctx context.Context, refundPayment entities.PaymentRefund) error
}

type ReceiptService interface {
	VoidReceipt(ctx context.Context, request entities.VoidReceipt) error
}
