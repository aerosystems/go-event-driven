package models

import (
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"time"
)

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
