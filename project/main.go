package main

import (
	"context"
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
	"net/http"
	"os"
)

type TicketsConfirmationRequest struct {
	Tickets []string `json:"tickets"`
}

func main() {
	log.Init(logrus.InfoLevel)
	clients, err := clients.NewClients(os.Getenv("GATEWAY_ADDR"), nil)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	receiptsClient := NewReceiptsClient(clients)
	spreadsheetsClient := NewSpreadsheetsClient(clients)

	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, log.NewWatermill(logrus.NewEntry(logrus.StandardLogger())))
	if err != nil {
		panic(err)
	}

	issueReceiptConsumer, err := NewConsumer(rdb, "issue-receipt-group")
	if err != nil {
		panic(err)
	}
	defer issueReceiptConsumer.Close()
	go func() {
		messages, err := issueReceiptConsumer.Subscribe("issue-receipt")
		if err != nil {
			panic(err)
		}

		for msg := range messages {
			ticketID := string(msg.Payload)
			if err := receiptsClient.IssueReceipt(context.Background(), ticketID); err != nil {
				logrus.Errorf("Error issuing receipt for ticket %s: %v", ticketID, err)
				msg.Nack()
				continue
			}
			msg.Ack()
		}

	}()
	appendTrackerConsumer, err := NewConsumer(rdb, "append-tracker-group")
	if err != nil {
		panic(err)
	}
	defer appendTrackerConsumer.Close()
	go func() {
		messages, err := appendTrackerConsumer.Subscribe("append-to-tracker")
		if err != nil {
			panic(err)
		}

		for msg := range messages {
			ticketID := string(msg.Payload)
			if err := spreadsheetsClient.AppendRow(context.Background(), "tickets-to-print", []string{ticketID}); err != nil {
				logrus.Errorf("Error appending row for ticket %s: %v", ticketID, err)
				msg.Nack()
				continue
			}
			msg.Ack()
		}
	}()

	e := commonHTTP.NewEcho()

	e.POST("/tickets-confirmation", func(c echo.Context) error {
		var request TicketsConfirmationRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}

		for _, ticket := range request.Tickets {
			if err := pub.Publish("issue-receipt", message.NewMessage(watermill.NewUUID(), []byte(ticket))); err != nil {
				return err
			}
			if err := pub.Publish("append-to-tracker", message.NewMessage(watermill.NewUUID(), []byte(ticket))); err != nil {
				return err
			}
		}

		return c.NoContent(http.StatusOK)
	})

	logrus.Info("Server starting...")

	err = e.Start(":8080")
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func (c ReceiptsClient) IssueReceipt(ctx context.Context, ticketID string) error {
	body := receipts.PutReceiptsJSONRequestBody{
		TicketId: ticketID,
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
