package main

import (
	"context"
	"os"
	"time"

	"github.com/Shopify/sarama"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

type PlayerJoined struct {
	PlayerID string    `json:"player_id"`
	JoinedAt time.Time `json:"joined_at"`
}

type PlayerLeft struct {
	PlayerID string    `json:"player_id"`
	LeftAt   time.Time `json:"left_at"`
}

type GameUpdated struct {
	Players []string `json:"players"`
}

func main() {
	logger := watermill.NewStdLogger(false, false)

	kafkaAddr := os.Getenv("KAFKA_ADDR")

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	eventProcessor, err := cqrs.NewEventGroupProcessorWithConfig(
		router,
		cqrs.EventGroupProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventGroupProcessorGenerateSubscribeTopicParams) (string, error) {
				return "events", nil
			},
			SubscriberConstructor: func(params cqrs.EventGroupProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return kafka.NewSubscriber(
					kafka.SubscriberConfig{
						Brokers:               []string{kafkaAddr},
						Unmarshaler:           kafka.DefaultMarshaler{},
						OverwriteSaramaConfig: newConfig(),
						ConsumerGroup:         "group",
					},
					logger,
				)
			},
			AckOnUnknownEvent: true,
			Marshaler:         cqrs.JSONMarshaler{},
			Logger:            logger,
		},
	)
	if err != nil {
		panic(err)
	}

	pub, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:   []string{kafkaAddr},
		Marshaler: kafka.DefaultMarshaler{},
	}, logger)
	if err != nil {
		panic(err)
	}

	eventBus, err := cqrs.NewEventBusWithConfig(
		pub,
		cqrs.EventBusConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return "events", nil
			},
			Marshaler: cqrs.JSONMarshaler{},
			Logger:    logger,
		},
	)
	if err != nil {
		panic(err)
	}

	gameHandler := &GameHandler{
		eventBus: eventBus,
	}

	eventProcessor.AddHandlersGroup(
		"orders",
		cqrs.NewGroupEventHandler(func(ctx context.Context, event *PlayerLeft) error {
			return gameHandler.HandlePlayerLeft(ctx, event)
		}),
		cqrs.NewGroupEventHandler(func(ctx context.Context, event *PlayerJoined) error {
			return gameHandler.HandlePlayerJoined(ctx, event)
		}),
	)

	err = router.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

type GameHandler struct {
	eventBus *cqrs.EventBus
	players  []string
}

func (h *GameHandler) HandlePlayerJoined(ctx context.Context, event *PlayerJoined) error {
	h.players = append(h.players, event.PlayerID)
	return h.publishGameUpdated(ctx)
}

func (h *GameHandler) HandlePlayerLeft(ctx context.Context, event *PlayerLeft) error {
	for i, player := range h.players {
		if player == event.PlayerID {
			h.players = append(h.players[:i], h.players[i+1:]...)
			break
		}
	}

	return h.publishGameUpdated(ctx)
}

func (h *GameHandler) publishGameUpdated(ctx context.Context) error {
	return h.eventBus.Publish(ctx, &GameUpdated{Players: h.players})
}

func newConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	return cfg
}
