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
		ConsumerGroup: "tickets-confirmation",
	}, logger)
	if err != nil {
		panic(err)
	}

	spreadsheetsSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "tickets-confirmation",
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

	router.AddNoPublisherHandler(
		"receipt-handler",
		"issue-receipt",
		receiptsSub,
		func(msg *message.Message) error {
			var payload IssueReceiptPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				return err
			}
			if err := receiptsClient.IssueReceipt(context.Background(), payload); err != nil {
				return fmt.Errorf("error issuing receipt for ticket %s: %v", payload.TicketID, err)
			}
			return nil
		})

	router.AddNoPublisherHandler(
		"spreadsheet-handler",
		"append-to-tracker",
		spreadsheetsSub,
		func(msg *message.Message) error {
			var payload AppendToTrackerPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				return err
			}
			if err := spreadsheetsClient.AppendRow(context.Background(), "tickets-to-print", []string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency}); err != nil {
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
		err := c.Bind(&request)
		if err != nil {
			return err
		}

		for _, ticket := range request.Tickets {
			if err := pub.Publish("issue-receipt", NewIssueReceiptMessage(ticket)); err != nil {
				return err
			}
			if err := pub.Publish("append-to-tracker", NewAppendToTrackerMessage(ticket)); err != nil {
				return err
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

func (c ReceiptsClient) IssueReceipt(ctx context.Context, request IssueReceiptPayload) error {
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

type IssueReceiptPayload struct {
	TicketID string       `json:"ticket_id"`
	Price    ReceiptPrice `json:"price"`
}
type AppendToTrackerPayload struct {
	TicketID      string       `json:"ticket_id"`
	CustomerEmail string       `json:"customer_email"`
	Price         ReceiptPrice `json:"price"`
}

type ReceiptPrice struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

func NewIssueReceiptMessage(ticket Ticket) *message.Message {
	issueReceipt := IssueReceiptPayload{
		TicketID: ticket.TicketID,
		Price: ReceiptPrice{
			Amount:   ticket.Price.Amount,
			Currency: ticket.Price.Currency,
		},
	}
	payload, err := json.Marshal(issueReceipt)
	if err != nil {
		panic(err)
	}
	return message.NewMessage(watermill.NewUUID(), payload)
}

func NewAppendToTrackerMessage(ticket Ticket) *message.Message {
	appendToTracker := AppendToTrackerPayload{
		TicketID:      ticket.TicketID,
		CustomerEmail: ticket.CustomerEmail,
		Price: ReceiptPrice{
			Amount:   ticket.Price.Amount,
			Currency: ticket.Price.Currency,
		},
	}
	payload, err := json.Marshal(appendToTracker)
	if err != nil {
		panic(err)
	}
	return message.NewMessage(watermill.NewUUID(), payload)
}
