package pubsub

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/sirupsen/logrus"
	"time"
)

func LoggingMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return message.HandlerFunc(func(message *message.Message) ([]*message.Message, error) {
		logger := log.FromContext(message.Context())
		logger.WithField("message_uuid", "message.UUID").Info("Handling a message")
		return next(message)
	})
}

func TracingMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return message.HandlerFunc(func(message *message.Message) ([]*message.Message, error) {
		correlationID := message.Metadata.Get("correlation_id")
		ctx := log.ToContext(message.Context(), logrus.WithFields(logrus.Fields{"correlation_id": correlationID}))
		message.SetContext(ctx)
		return next(message)
	})
}

func HandleErrorsMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return message.HandlerFunc(func(message *message.Message) ([]*message.Message, error) {
		messages, err := next(message)
		if err != nil {
			logger := log.FromContext(message.Context())
			logger.WithField("error", err.Error()).WithField("message_uuid", message.UUID).Error("Message handling error")
		}
		return messages, nil
	})
}

func ExponentialBackoffMiddleware(next message.HandlerFunc) message.HandlerFunc {
	retry := middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          watermill.NewStdLogger(false, false),
	}
	return retry.Middleware(func(message *message.Message) ([]*message.Message, error) {
		return next(message)
	})
}
