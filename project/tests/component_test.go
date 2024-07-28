package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
	"github.com/redis/go-redis/v9"
	"net/http"
	"os"
	"os/signal"
	"testing"
	"tickets/config"
	"tickets/infra/adapters"
	"tickets/models"
	"tickets/service"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	cfg := config.NewConfig()
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddress})
	defer redisClient.Close()

	spreadsheetsClientMock := &adapters.SpreadsheetsClientMock{}
	receiptsClientMock := &adapters.ReceiptsClientMock{}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	go func() {
		svc := service.NewService(redisClient, spreadsheetsClientMock, receiptsClientMock)
		assert.NoError(t, svc.Run(ctx))
	}()

	waitForHttpServer(t)

	ticket := TicketStatus{
		TicketID: uuid.NewString(),
		Status:   "confirmed",
		Price: Money{
			Amount:   "50.30",
			Currency: "GBP",
		},
		Email:     "email@example.com",
		BookingID: uuid.NewString(),
	}

	sendTicketsStatus(t, TicketsStatusRequest{[]TicketStatus{ticket}})

	assertReceiptForTicketIssued(t, receiptsClientMock, ticket)
	assertSpreadsheetAppendRowForTicket(t, spreadsheetsClientMock, ticket, "tickets-to-print")

	sendTicketsStatus(t, TicketsStatusRequest{
		[]TicketStatus{
			{
				TicketID: ticket.TicketID,
				Status:   "canceled",
				Email:    ticket.Email,
			},
		},
	})

	assertSpreadsheetAppendRowForTicket(t, spreadsheetsClientMock, ticket, "tickets-to-refund")
}

func waitForHttpServer(t *testing.T) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			resp, err := http.Get("http://localhost:8080/health")
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			if assert.Less(t, resp.StatusCode, 300, "API not ready, http status: %d", resp.StatusCode) {
				return
			}
		},
		time.Second*10,
		time.Millisecond*50,
	)
}

type TicketsStatusRequest struct {
	Tickets []TicketStatus `json:"tickets"`
}

type TicketStatus struct {
	TicketID  string `json:"ticket_id"`
	Status    string `json:"status"`
	Price     Money  `json:"price"`
	Email     string `json:"email"`
	BookingID string `json:"booking_id"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

func sendTicketsStatus(t *testing.T, req TicketsStatusRequest) {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	correlationID := shortuuid.New()

	ticketIDs := make([]string, 0, len(req.Tickets))
	for _, ticket := range req.Tickets {
		ticketIDs = append(ticketIDs, ticket.TicketID)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/tickets-status",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set("Correlation-ID", correlationID)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func assertReceiptForTicketIssued(t *testing.T, receiptsClient *adapters.ReceiptsClientMock, ticket TicketStatus) {
	assert.EventuallyWithT(
		t,
		func(collectT *assert.CollectT) {
			issuedReceipts := len(receiptsClient.IssuedReceipts)
			t.Log("issued receipts", issuedReceipts)

			assert.Greater(collectT, issuedReceipts, 0, "no receipts issued")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var receipt models.IssueReceiptRequest
	var ok bool
	for _, issuedReceipt := range receiptsClient.IssuedReceipts {
		if issuedReceipt.TicketID != ticket.TicketID {
			continue
		}
		receipt = issuedReceipt
		ok = true
		break
	}

	require.Truef(t, ok, "receipt for ticket %s not found", ticket.TicketID)
	assert.Equal(t, ticket.TicketID, receipt.TicketID)
	assert.Equal(t, ticket.Price.Amount, receipt.Price.Amount)
	assert.Equal(t, ticket.Price.Currency, receipt.Price.Currency)
}

func assertSpreadsheetAppendRowForTicket(t *testing.T, spreadsheetsClient *adapters.SpreadsheetsClientMock, ticket TicketStatus, spreadsheetName string) {
	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			rows, ok := spreadsheetsClient.AppendedRows[spreadsheetName]
			if !assert.True(t, ok, "sheet %s not found", spreadsheetName) {
				return
			}

			var allValues []string

			for _, row := range rows {
				for _, col := range row {
					allValues = append(allValues, col)
				}
			}

			assert.Contains(t, allValues, ticket.TicketID, "ticket id not found in sheet %s", spreadsheetName)
		},
		10*time.Second,
		100*time.Millisecond,
	)
}
