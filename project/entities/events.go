package entities

import (
	"time"

	"github.com/google/uuid"
)

type EventHeader struct {
	ID             string    `json:"id"`
	PublishedAt    time.Time `json:"published_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

func NewEventHeader() EventHeader {
	return EventHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: uuid.NewString(),
	}
}

type TicketBookingConfirmed struct {
	Header EventHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`

	BookingID string `json:"booking_id"`
}

type TicketBookingCanceled struct {
	Header EventHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}

type TicketPrinted struct {
	Header EventHeader `json:"header"`

	TicketID string `json:"ticket_id"`
	FileName string `json:"file_name"`
}

type TicketRefunded struct {
	Header EventHeader `json:"header"`

	TicketID string `json:"ticket_id"`
}

type BookingMade struct {
	Header EventHeader `json:"header"`

	NumberOfTickets int `json:"number_of_tickets"`

	BookingID uuid.UUID `json:"booking_id"`

	CustomerEmail string    `json:"customer_email"`
	ShowId        uuid.UUID `json:"show_id"`
}

type DataLakeEvent struct {
	ID          string    `json:"event_id"`
	Name        string    `json:"event_name"`
	Payload     []byte    `json:"event_payload"`
	PublishedAt time.Time `json:"published_at"`
}
