package http

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h Handler) OpsGetBookings(c echo.Context) error {
	bookings, err := h.opsReadModel.AllReservations()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, bookings)
}

func (h Handler) OpsGetBooking(c echo.Context) error {
	booking, err := h.opsReadModel.ReservationReadModel(c.Request().Context(), c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, booking)
}
