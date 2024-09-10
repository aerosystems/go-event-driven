package command

import "github.com/ThreeDotsLabs/watermill/components/cqrs"

type Handler struct {
	commandBus *cqrs.CommandBus
}

func NewHandler(commandBus *cqrs.CommandBus) Handler {
	if commandBus == nil {
		panic("missing commandBus")
	}

	return Handler{
		commandBus: commandBus,
	}
}
