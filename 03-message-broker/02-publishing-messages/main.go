package main

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"os"
)

func main() {
	logger := watermill.NewStdLogger(false, false)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}

	messages := []*message.Message{
		message.NewMessage(watermill.NewUUID(), []byte("50")),
		message.NewMessage(watermill.NewUUID(), []byte("100")),
	}
	for _, msg := range messages {
		err := pub.Publish("progress", msg)
		if err != nil {
			panic(err)
			return
		}
	}
}
