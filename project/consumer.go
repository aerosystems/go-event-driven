package main

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	rdb *redis.Client
	sub *redisstream.Subscriber
}

func NewConsumer(rdb *redis.Client, consumerGroup string) (*Consumer, error) {
	logger := log.NewWatermill(logrus.NewEntry(logrus.StandardLogger()))
	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: consumerGroup,
	}, logger)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		rdb: rdb,
		sub: sub,
	}, nil
}

func (c Consumer) Subscribe(topic string) (<-chan *message.Message, error) {
	messages, err := c.sub.Subscribe(context.Background(), topic)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (c Consumer) Close() error {
	return c.sub.Close()
}
