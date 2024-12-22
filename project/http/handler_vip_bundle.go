package http

import (
	"net/http"
	"tickets/entities"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type vipBundleRequest struct {
	CustomerEmail   string    `json:"customer_email"`
	InboundFlightId uuid.UUID `json:"inbound_flight_id"`
	NumberOfTickets int       `json:"number_of_tickets"`
	Passengers      []string  `json:"passengers"`
	ReturnFlightId  uuid.UUID `json:"return_flight_id"`
	ShowId          uuid.UUID `json:"show_id"`
}

type vipBundleResponse struct {
	BookingId   uuid.UUID `json:"booking_id"`
	VipBundleId uuid.UUID `json:"vip_bundle_id"`
}

func (h Handler) PostVipBundle(c echo.Context) error {
	var request vipBundleRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	if request.NumberOfTickets < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "number of tickets must be greater than 0")
	}

	vb := entities.VipBundle{
		VipBundleID:     uuid.New(),
		BookingID:       uuid.New(),
		CustomerEmail:   request.CustomerEmail,
		NumberOfTickets: request.NumberOfTickets,
		ShowId:          request.ShowId,
		Passengers:      request.Passengers,
		InboundFlightID: request.InboundFlightId,
		ReturnFlightID:  request.ReturnFlightId,
		IsFinalized:     false,
		Failed:          false,
	}

	if err := h.vipBundlesRepository.Add(c.Request().Context(), vb); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, vipBundleResponse{
		BookingId:   vb.BookingID,
		VipBundleId: vb.VipBundleID,
	})
}
