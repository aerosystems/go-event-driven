package HttpTicketHandler

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"tickets/models"
)

type TicketsStatusRequest struct {
	Tickets []models.Ticket `json:"tickets"`
}

func (h Handler) TicketsStatus(c echo.Context) error {
	var request TicketsStatusRequest
	if err := c.Bind(&request); err != nil {
		return err
	}

	for _, ticket := range request.Tickets {
		switch ticket.Status {
		case "confirmed":
			if err := h.ticketPub.Publish("TicketBookingConfirmed", models.NewTicketBookingConfirmedMessage(ticket, c.Request().Header.Get("Correlation-ID"))); err != nil {
				return err
			}
		case "canceled":
			if err := h.ticketPub.Publish("TicketBookingCanceled", models.NewTicketBookingCanceledMessage(ticket, c.Request().Header.Get("Correlation-ID"))); err != nil {
				return err
			}
		}
	}

	return c.NoContent(http.StatusOK)
}
