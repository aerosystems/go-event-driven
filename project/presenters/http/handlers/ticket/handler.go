package HttpTicketHandler

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	eventBus *cqrs.EventBus
}

func NewHttpTicketHandler(eventBus *cqrs.EventBus) *Handler {
	return &Handler{
		eventBus,
	}
}
