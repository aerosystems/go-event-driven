package http

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"tickets/entities"
)

type ticketsStatusRequest struct {
	Tickets []ticketStatusRequest `json:"tickets"`
}

type ticketStatusRequest struct {
	TicketID      string         `json:"ticket_id"`
	Status        string         `json:"status"`
	Price         entities.Money `json:"price"`
	CustomerEmail string         `json:"customer_email"`
	BookingID     string         `json:"booking_id"`
}

func (h Handler) PostTicketsStatus(c echo.Context) error {
	var request ticketsStatusRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	idempotencyKey := c.Request().Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Idempotency-Key header is required")
	}

	for _, ticket := range request.Tickets {
		if ticket.Status == "confirmed" {
			event := entities.TicketBookingConfirmed_v1{
				Header: entities.NewEventHeaderWithIdempotencyKey(idempotencyKey + ticket.TicketID),

				TicketID:      ticket.TicketID,
				Price:         ticket.Price,
				CustomerEmail: ticket.CustomerEmail,

				BookingID: ticket.BookingID,
			}

			if err := h.eventBus.Publish(c.Request().Context(), event); err != nil {
				return fmt.Errorf("failed to publish TicketBookingConfirmed event: %w", err)
			}
		} else if ticket.Status == "canceled" {
			event := entities.TicketBookingCanceled_v1{
				Header:        entities.NewEventHeaderWithIdempotencyKey(idempotencyKey + ticket.TicketID),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}

			if err := h.eventBus.Publish(c.Request().Context(), event); err != nil {
				return fmt.Errorf("failed to publish TicketBookingCanceled event: %w", err)
			}
		} else {
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h Handler) PutTicketRefund(c echo.Context) error {
	ticketID := c.Param("ticket_id")

	event := entities.RefundTicket{
		Header:   entities.NewEventHeaderWithIdempotencyKey(uuid.NewString()),
		TicketID: ticketID,
	}

	if err := h.commandBus.Send(c.Request().Context(), event); err != nil {
		return fmt.Errorf("failed to send RefundTicket command: %w", err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h Handler) GetTickets(c echo.Context) error {
	tickets, err := h.ticketsRepo.FindAll(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to find tickets: %w", err)
	}

	return c.JSON(http.StatusOK, tickets)
}
