package PubSubRouter

import (
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func (r *Router) RegisterEventHandlers(sub message.Subscriber, handlers []cqrs.EventHandler, watermillLogger watermill.LoggerAdapter) error {
	ep, err := cqrs.NewEventProcessorWithConfig(
		r.router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return fmt.Sprintf("svc-tickets-%s", params.EventName), nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return sub, nil
			},
			Logger: watermillLogger,
		})
	if err == nil {
		panic(err)
	}
	return ep.AddHandlers(handlers...)
}
