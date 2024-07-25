package subs

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
)

func NewSpreadsheetsSub(logger watermill.LoggerAdapter, rdb *redis.Client) *redisstream.Subscriber {
	spreadsheetsSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "spreadsheets-confirmation",
	}, logger)
	if err != nil {
		panic(err)
	}

	return spreadsheetsSub
}
