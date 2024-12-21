package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"log"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/urfave/cli/v2"
)

const PoisonQueueTopic = "PoisonQueue"

type Message struct {
	ID     string
	Reason string
}

type Handler struct {
	subscriber message.Subscriber
	publisher  message.Publisher
}

func NewHandler() (*Handler, error) {
	logger := watermill.NewStdLogger(false, false)

	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	sub, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               []string{os.Getenv("KAFKA_ADDR")},
			Unmarshaler:           kafka.DefaultMarshaler{},
			ConsumerGroup:         "poison-queue-cli",
			OverwriteSaramaConfig: cfg,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	pub, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   []string{os.Getenv("KAFKA_ADDR")},
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &Handler{
		subscriber: sub,
		publisher:  pub,
	}, nil
}

func (h *Handler) Preview(ctx context.Context) ([]Message, error) {
	var result []Message

	router, err := message.NewRouter(
		message.RouterConfig{},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	firstMessageUUID := ""

	done := false

	router.AddHandler(
		"preview",
		PoisonQueueTopic,
		h.subscriber,
		PoisonQueueTopic,
		h.publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			if done {
				cancel()
				return nil, errors.New("done")
			}

			if firstMessageUUID == "" {
				firstMessageUUID = msg.UUID
			} else if firstMessageUUID == msg.UUID {
				// we've read all messages
				done = true
				return nil, errors.New("done")
			}

			result = append(result, Message{
				ID:     msg.UUID,
				Reason: msg.Metadata.Get(middleware.ReasonForPoisonedKey),
			})

			return []*message.Message{msg}, nil
		},
	)

	err = router.Run(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func main() {
	app := &cli.App{
		Name:  "poison-queue-cli",
		Usage: "Manage the Poison Queue",
		Commands: []*cli.Command{
			{
				Name:  "preview",
				Usage: "preview messages",
				Action: func(c *cli.Context) error {
					h, err := NewHandler()
					if err != nil {
						return err
					}

					messages, err := h.Preview(c.Context)
					if err != nil {
						return err
					}

					for _, m := range messages {
						fmt.Printf("%v\t%v\n", m.ID, m.Reason)
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
