package entities

import (
	"github.com/google/uuid"
	"time"
)

type CommandHeader struct {
	ID             string    `json:"id"`
	PublishedAt    time.Time `json:"published_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

func NewCommandHeader() CommandHeader {
	return CommandHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: uuid.NewString(),
	}
}

type RefundTicket struct {
	Header CommandHeader `json:"header"`

	TicketID string `json:"ticket_id"`
}
