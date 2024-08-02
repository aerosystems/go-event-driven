package models

import (
	"github.com/ThreeDotsLabs/watermill"
	"time"
)

type TicketBookingCanceled struct {
	Header        EventHeader `json:"header"`
	TicketID      string      `json:"ticket_id"`
	CustomerEmail string      `json:"customer_email"`
	Price         Money       `json:"price"`
}

func NewTicketBookingCanceledMessage(ticket Ticket) TicketBookingCanceled {
	return TicketBookingCanceled{
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
}
