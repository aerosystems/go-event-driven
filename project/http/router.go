package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
)

func NewHttpRouter(
	eventBus *cqrs.EventBus,
	spreadsheetsAPIClient SpreadsheetsAPI,
	ticketRepo TicketRepository,
	showRepo ShowRepository,
	bookingRepo BookingRepository,
) *echo.Echo {
	e := libHttp.NewEcho()

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	handler := Handler{
		eventBus:              eventBus,
		spreadsheetsAPIClient: spreadsheetsAPIClient,
		ticketRepo:            ticketRepo,
		showRepo:              showRepo,
		bookingRepo:           bookingRepo,
	}

	e.POST("/tickets-status", handler.PostTicketsStatus)
	e.GET("/tickets", handler.GetAllTickets)
	e.POST("/shows", handler.PostShow)
	e.POST("/book-tickets", handler.PostBookTickets)

	return e
}
