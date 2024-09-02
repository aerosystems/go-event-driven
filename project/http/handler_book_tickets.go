package http

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"tickets/entities"
)

type BookTicketRequest struct {
	ShowID          string `json:"show_id"`
	NumberOfTickets int    `json:"number_of_tickets"`
	CustomerEmail   string `json:"customer_email"`
}

type BookTicketResponse struct {
	BookingID string `json:"booking_id"`
}

func (h Handler) PostBookTicket(c echo.Context) error {
	ctx := c.Request().Context()

	var req BookTicketRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	booking := entities.Booking{
		BookingID:       uuid.New().String(),
		ShowID:          req.ShowID,
		NumberOfTickets: req.NumberOfTickets,
		CustomerEmail:   req.CustomerEmail,
	}

	bookingID, err := h.bookingRepo.Create(ctx, booking)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, BookTicketResponse{BookingID: bookingID})
}
