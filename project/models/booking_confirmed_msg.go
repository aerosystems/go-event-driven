package models

import (
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"time"
)

type EventHeader struct {
	ID          string `json:"id"`
	PublishedAt string `json:"published_at"`
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
