package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/receipts"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/spreadsheets"
	commonHTTP "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const brokenMsgUuid = "2beaf5bc-d5e4-4653-b075-2b36bbf28949"

type Ticket struct {
	TicketID      string      `json:"ticket_id"`
	Status        string      `json:"status"`
	CustomerEmail string      `json:"customer_email"`
	Price         TicketPrice `json:"price"`
}

type TicketPrice struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type TicketsStatusRequest struct {
	Tickets []Ticket `json:"tickets"`
}

func main() {
	log.Init(logrus.InfoLevel)
	logger := watermill.NewStdLogger(false, false)

	clients, err := clients.NewClients(
		os.Getenv("GATEWAY_ADDR"),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	receiptsClient := NewReceiptsClient(clients)
	spreadsheetsClient := NewSpreadsheetsClient(clients)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	receiptsSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "receipts-confirmation",
	}, logger)
	if err != nil {
		panic(err)
	}

	spreadsheetsSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "spreadsheets-confirmation",
	}, logger)
	if err != nil {
		panic(err)
	}

	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, log.NewWatermill(logrus.NewEntry(logrus.StandardLogger())))
	if err != nil {
		panic(err)
	}

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}
	router.AddMiddleware(TracingMiddleware, LoggingMiddleware, HandleErrorsMiddleware, ExponentialBackoffMiddleware)

	router.AddNoPublisherHandler(
		"receipt-handler",
		"TicketBookingConfirmed",
		receiptsSub,
		func(msg *message.Message) error {
			if msg.UUID == brokenMsgUuid {
				return nil
			}
			if msg.Metadata.Get("type") != "TicketBookingConfirmed" {
				return nil
			}
			var payload TicketBookingConfirmed
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				return err
			}
			if payload.Price.Currency == "" {
				payload.Price.Currency = "USD"
			}
			reqCorrelationID := msg.Metadata.Get("correlation_id")
			ctx := log.ContextWithCorrelationID(msg.Context(), reqCorrelationID)
			msg.SetContext(ctx)
			if err := receiptsClient.IssueReceipt(msg.Context(), IssueReceiptRequest{
				TicketID: payload.TicketID,
				Price:    payload.Price,
			}); err != nil {
				return err
			}
			return nil
		})

	router.AddNoPublisherHandler(
		"spreadsheet-confirmed-handler",
		"TicketBookingConfirmed",
		spreadsheetsSub,
		func(msg *message.Message) error {
			if msg.UUID == brokenMsgUuid {
				return nil
			}
			if msg.Metadata.Get("type") != "TicketBookingConfirmed" {
				return nil
			}
			var payload TicketBookingConfirmed
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				return err
			}
			if payload.Price.Currency == "" {
				payload.Price.Currency = "USD"
			}
			reqCorrelationID := msg.Metadata.Get("correlation_id")
			ctx := log.ContextWithCorrelationID(msg.Context(), reqCorrelationID)
			msg.SetContext(ctx)
			if err := spreadsheetsClient.AppendRow(msg.Context(), "tickets-to-print", []string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency}); err != nil {
				return err
			}
			return nil
		})

	router.AddNoPublisherHandler(
		"spreadsheet-canceled-handler",
		"TicketBookingCanceled",
		spreadsheetsSub,
		func(msg *message.Message) error {
			if msg.UUID == brokenMsgUuid {
				return nil
			}
			if msg.Metadata.Get("type") != "TicketBookingCanceled" {
				return nil
			}
			var payload TicketBookingCanceled
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				return err
			}
			reqCorrelationID := msg.Metadata.Get("correlation_id")
			ctx := log.ContextWithCorrelationID(msg.Context(), reqCorrelationID)
			msg.SetContext(ctx)
			if err := spreadsheetsClient.AppendRow(msg.Context(), "tickets-to-refund", []string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency}); err != nil {
				return err
			}
			return nil
		})

	e := commonHTTP.NewEcho()

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	e.POST("/tickets-status", func(c echo.Context) error {
		var request TicketsStatusRequest
		if err := c.Bind(&request); err != nil {
			return err
		}

		for _, ticket := range request.Tickets {
			switch ticket.Status {
			case "confirmed":
				if err := pub.Publish("TicketBookingConfirmed", NewTicketBookingConfirmedMessage(ticket, c.Request().Header.Get("Correlation-ID"))); err != nil {
					return err
				}
			case "canceled":
				if err := pub.Publish("TicketBookingCanceled", NewTicketBookingCanceledMessage(ticket, c.Request().Header.Get("Correlation-ID"))); err != nil {
					return err
				}
			}
		}

		return c.NoContent(http.StatusOK)
	})

	logrus.Info("Server starting...")

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return router.Run(context.Background())
	})
	g.Go(func() error {
		<-router.Running()
		err := e.Start(":8080")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	g.Go(func() error {
		<-ctx.Done()
		return e.Shutdown(ctx)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}

type ReceiptsClient struct {
	clients *clients.Clients
}

func NewReceiptsClient(clients *clients.Clients) ReceiptsClient {
	return ReceiptsClient{
		clients: clients,
	}
}

type IssueReceiptRequest struct {
	TicketID string
	Price    Money
}

func (c ReceiptsClient) IssueReceipt(ctx context.Context, request IssueReceiptRequest) error {
	body := receipts.PutReceiptsJSONRequestBody{
		TicketId: request.TicketID,
		Price: receipts.Money{
			MoneyAmount:   request.Price.Amount,
			MoneyCurrency: request.Price.Currency,
		},
	}

	receiptsResp, err := c.clients.Receipts.PutReceiptsWithResponse(ctx, body)
	if err != nil {
		return err
	}
	if receiptsResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", receiptsResp.StatusCode())
	}

	return nil
}

type SpreadsheetsClient struct {
	clients *clients.Clients
}

func NewSpreadsheetsClient(clients *clients.Clients) SpreadsheetsClient {
	return SpreadsheetsClient{
		clients: clients,
	}
}

func (c SpreadsheetsClient) AppendRow(ctx context.Context, spreadsheetName string, row []string) error {
	request := spreadsheets.PostSheetsSheetRowsJSONRequestBody{
		Columns: row,
	}

	sheetsResp, err := c.clients.Spreadsheets.PostSheetsSheetRowsWithResponse(ctx, spreadsheetName, request)
	if err != nil {
		return err
	}
	if sheetsResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", sheetsResp.StatusCode())
	}

	return nil
}

type EventHeader struct {
	ID          string `json:"id"`
	PublishedAt string `json:"published_at"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type TicketBookingConfirmed struct {
	Header        EventHeader `json:"header"`
	TicketID      string      `json:"ticket_id"`
	CustomerEmail string      `json:"customer_email"`
	Price         Money       `json:"price"`
}

func NewTicketBookingConfirmedMessage(ticket Ticket, correlationId string) *message.Message {
	ticketBookingConfirmed := TicketBookingConfirmed{
		Header: EventHeader{
			ID:          watermill.NewUUID(),
			PublishedAt: time.Now().Format(time.RFC3339),
		},
		TicketID:      ticket.TicketID,
		CustomerEmail: ticket.CustomerEmail,
		Price: Money{
			Amount:   ticket.Price.Amount,
			Currency: ticket.Price.Currency,
		},
	}

	payload, err := json.Marshal(ticketBookingConfirmed)
	if err != nil {
		panic(err)
	}

	msg := message.NewMessage(ticketBookingConfirmed.Header.ID, payload)
	if correlationId != "" {
		msg.Metadata.Set("correlation_id", correlationId)
	}
	msg.Metadata.Set("type", "TicketBookingConfirmed")
	return msg
}

type TicketBookingCanceled struct {
	Header        EventHeader `json:"header"`
	TicketID      string      `json:"ticket_id"`
	CustomerEmail string      `json:"customer_email"`
	Price         Money       `json:"price"`
}

func NewTicketBookingCanceledMessage(ticket Ticket, correlationId string) *message.Message {
	ticketBookingCanceled := TicketBookingCanceled{
		Header: EventHeader{
			ID:          watermill.NewUUID(),
			PublishedAt: time.Now().Format(time.RFC3339),
		},
		TicketID:      ticket.TicketID,
		CustomerEmail: ticket.CustomerEmail,
		Price: Money{
			Amount:   ticket.Price.Amount,
			Currency: ticket.Price.Currency,
		},
	}

	payload, err := json.Marshal(ticketBookingCanceled)
	if err != nil {
		panic(err)
	}

	msg := message.NewMessage(ticketBookingCanceled.Header.ID, payload)
	if correlationId != "" {
		msg.Metadata.Set("correlation_id", correlationId)
	}
	msg.Metadata.Set("type", "TicketBookingCanceled")
	return msg
}

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
