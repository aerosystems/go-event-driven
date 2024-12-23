package message

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/lithammer/shortuuid/v3"
	"github.com/sirupsen/logrus"
)

var (
	messagesProcessedTotalCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "messages",
		Name:      "processed_total",
	}, []string{"topic", "handler"})

	messagesProcessingFailedTotalCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "messages",
		Name:      "processing_failed_total",
	}, []string{"topic", "handler"})

	messagesProcessingDurationSeconds = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  "messages",
		Name:       "processing_duration_seconds",
		Help:       "The total time spent processing messages",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"topic", "handler"})
)

func useMiddlewares(router *message.Router, watermillLogger watermill.LoggerAdapter) {
	router.AddMiddleware(middleware.Recoverer)

	router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          watermillLogger,
	}.Middleware)

	router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) (events []*message.Message, err error) {
			ctx := msg.Context()

			reqCorrelationID := msg.Metadata.Get("correlation_id")
			if reqCorrelationID == "" {
				reqCorrelationID = shortuuid.New()
			}

			ctx = log.ToContext(ctx, logrus.WithFields(logrus.Fields{"correlation_id": reqCorrelationID}))
			ctx = log.ContextWithCorrelationID(ctx, reqCorrelationID)

			msg.SetContext(ctx)

			return h(msg)
		}
	})

	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			logger := log.FromContext(msg.Context()).WithFields(logrus.Fields{
				"message_id": msg.UUID,
				"payload":    string(msg.Payload),
				"metadata":   msg.Metadata,
			})

			logger.Info("Handling a message")

			msgs, err := next(msg)
			if err != nil {
				logger.WithError(err).Error("Error while handling a message")
			}

			return msgs, err
		}
	})

	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			subscribeTopicName := message.SubscriberNameFromCtx(msg.Context())
			handlerName := message.HandlerNameFromCtx(msg.Context())

			labels := prometheus.Labels{"topic": subscribeTopicName, "handler": handlerName}
			messagesProcessedTotalCounter.With(labels).Inc()

			timeStart := time.Now()
			msgs, err := next(msg)
			if err != nil {
				messagesProcessingFailedTotalCounter.With(labels).Inc()
				return msgs, err
			}

			messagesProcessingDurationSeconds.With(labels).Observe(time.Since(timeStart).Seconds())

			return msgs, nil

		}
	})

	router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) (events []*message.Message, err error) {
			topic := message.SubscribeTopicFromCtx(msg.Context())
			handler := message.HandlerNameFromCtx(msg.Context())

			ctx := msg.Context()

			ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(msg.Metadata))

			ctx, span := otel.Tracer("").Start(
				ctx,
				fmt.Sprintf("topic: %s, handler: %s", topic, handler),
				trace.WithAttributes(
					attribute.String("topic", topic),
					attribute.String("handler", handler),
				),
			)
			defer span.End()

			msg.SetContext(ctx)

			msgs, err := h(msg)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}

			return msgs, err
		}
	})
}
