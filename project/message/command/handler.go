package command

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/entities"
)

type Handler struct {
	commandBus     *cqrs.CommandBus
	refundsService RefundsService
}

func NewHandler(commandBus *cqrs.CommandBus, receiptsService RefundsService) Handler {
	if commandBus == nil {
		panic("missing commandBus")
	}
	if receiptsService == nil {
		panic("missing refundsService")
	}

	return Handler{
		commandBus:     commandBus,
		refundsService: receiptsService,
	}
}

type RefundsService interface {
	RefundReceipt(ctx context.Context, request entities.VoidReceipt) error
	VoidReceipt(ctx context.Context, request entities.VoidReceipt) error
}
