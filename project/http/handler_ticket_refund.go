package http

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"tickets/entities"
)

func (h Handler) PutTicketRefund(c echo.Context) error {
	ticketID := c.Param("ticket_id")

	if err := h.commandBus.Send(context.Background(), entities.RefundTicket{
		TicketID: ticketID,
	}); err != nil {
		return err
	}
	return c.NoContent(http.StatusAccepted)
}
