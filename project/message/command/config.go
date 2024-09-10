package command

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

var marshaler = cqrs.JSONMarshaler{
	GenerateName: cqrs.StructName,
}

func NewProcessorConfig(redisClient *redis.Client, watermillLogger watermill.LoggerAdapter) cqrs.CommandProcessorConfig {
	return cqrs.CommandProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
			return params.CommandName, nil
		},
		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return redisstream.NewSubscriber(redisstream.SubscriberConfig{
				Client:        redisClient,
				ConsumerGroup: "svc-tickets." + params.HandlerName,
			}, watermillLogger)
		},
		Marshaler: marshaler,
	}
}
