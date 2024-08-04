package http

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"tickets/entities"
)

type Ticket struct {
	ID            string      `json:"ticket_id"`
	CustomerEmail string      `json:"customer_email"`
	TicketPrice   TicketPrice `json:"price"`
}

type TicketPrice struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

func entitiesToTickets(tickets []entities.Ticket) []Ticket {
	result := make([]Ticket, 0, len(tickets))
	for _, ticket := range tickets {
		result = append(result, Ticket{
			ID:            ticket.TicketID,
			CustomerEmail: ticket.CustomerEmail,
			TicketPrice: TicketPrice{
				Amount:   ticket.Price.Amount,
				Currency: ticket.Price.Currency,
			},
		})
	}
	return result
}
func (h Handler) GetAllTickets(c echo.Context) error {
	ctx := context.Background()
	tickets, err := h.ticketRepo.GetAll(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, entitiesToTickets(tickets))
}
