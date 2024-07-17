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
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
)

type TicketsConfirmationRequest struct {
	Tickets []string `json:"tickets"`
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

	generalPub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
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
			ticketID := string(msg.Payload)
			if err := receiptsClient.IssueReceipt(context.Background(), ticketID); err != nil {
				return fmt.Errorf("error issuing receipt for ticket %s: %v", ticketID, err)
			}
			return nil
		})

	router.AddNoPublisherHandler(
		"spreadsheet-handler",
		"append-to-tracker",
		spreadsheetsSub,
		func(msg *message.Message) error {
			ticketID := string(msg.Payload)
			if err := spreadsheetsClient.AppendRow(context.Background(), "tickets-to-print", []string{ticketID}); err != nil {
				return fmt.Errorf("error appending row for ticket %s: %v", ticketID, err)
			}
			return nil
		})

	e := commonHTTP.NewEcho()

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	e.POST("/tickets-confirmation", func(c echo.Context) error {
		var request TicketsConfirmationRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}

		for _, ticket := range request.Tickets {
			msg := message.NewMessage(watermill.NewUUID(), []byte(ticket))
			if err := generalPub.Publish("issue-receipt", msg); err != nil {
				return err
			}
			if err := generalPub.Publish("append-to-tracker", msg); err != nil {
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
