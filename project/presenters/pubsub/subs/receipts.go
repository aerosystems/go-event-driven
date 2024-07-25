package subs

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
)

func NewReceiptsSub(logger watermill.LoggerAdapter, rdb *redis.Client) *redisstream.Subscriber {
	receiptsSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "receipts-confirmation",
	}, logger)
	if err != nil {
		panic(err)
	}

	return receiptsSub
}
