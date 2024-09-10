package command

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewBus(pub message.Publisher) *cqrs.CommandBus {
	commandBus, err := cqrs.NewCommandBusWithConfig(
		pub,
		cqrs.CommandBusConfig{
			GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
				return params.CommandName, nil
			},
			Marshaler: marshaler,
		},
	)
	if err != nil {
		panic(err)
	}

	return commandBus
}
