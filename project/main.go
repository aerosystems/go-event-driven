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
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"time"
)

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

	clients, err := clients.NewClients(os.Getenv("GATEWAY_ADDR"), nil)
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
	router.AddMiddleware(LoggingMiddleware)

	router.AddNoPublisherHandler(
		"receipt-handler",
		"TicketBookingConfirmed",
		receiptsSub,
		func(msg *message.Message) error {
			var payload TicketBookingConfirmed
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				return err
			}
			if err := receiptsClient.IssueReceipt(context.Background(), IssueReceiptRequest{
				TicketID: payload.TicketID,
				Price:    payload.Price,
			}); err != nil {
				return fmt.Errorf("error issuing receipt for ticket %s: %v", payload.TicketID, err)
			}
			return nil
		})

	router.AddNoPublisherHandler(
		"spreadsheet-confirmed-handler",
		"TicketBookingConfirmed",
		spreadsheetsSub,
		func(msg *message.Message) error {
			var payload TicketBookingConfirmed
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				return err
			}
			if err := spreadsheetsClient.AppendRow(context.Background(), "tickets-to-print", []string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency}); err != nil {
				return fmt.Errorf("error appending row for ticket %s: %v", payload.TicketID, err)
			}
			return nil
		})

	router.AddNoPublisherHandler(
		"spreadsheet-canceled-handler",
		"TicketBookingCanceled",
		spreadsheetsSub,
		func(msg *message.Message) error {
			var payload TicketBookingCanceled
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				return err
			}
			if err := spreadsheetsClient.AppendRow(context.Background(), "tickets-to-refund", []string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency}); err != nil {
				return fmt.Errorf("error appending row for ticket %s: %v", payload.TicketID, err)
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
				if err := pub.Publish("TicketBookingConfirmed", NewTicketBookingConfirmedMessage(ticket)); err != nil {
					return err
				}
			case "canceled":
				if err := pub.Publish("TicketBookingCanceled", NewTicketBookingCanceledMessage(ticket)); err != nil {
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

func NewTicketBookingConfirmedMessage(ticket Ticket) *message.Message {
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

	return message.NewMessage(ticketBookingConfirmed.Header.ID, payload)
}

type TicketBookingCanceled struct {
	Header        EventHeader `json:"header"`
	TicketID      string      `json:"ticket_id"`
	CustomerEmail string      `json:"customer_email"`
	Price         Money       `json:"price"`
}

func NewTicketBookingCanceledMessage(ticket Ticket) *message.Message {
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

	return message.NewMessage(ticketBookingCanceled.Header.ID, payload)
}

func LoggingMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return message.HandlerFunc(func(message *message.Message) ([]*message.Message, error) {
		logrus.WithField("message_uuid", message.UUID).Info("Handling a message")
		return next(message)
	})
}
