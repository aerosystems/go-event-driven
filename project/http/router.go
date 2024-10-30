package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
)

func NewHttpRouter(
	commandBus *cqrs.CommandBus,
	eventBus *cqrs.EventBus,
	spreadsheetsAPIClient SpreadsheetsAPI,
	ticketRepo TicketRepository,
	showRepo ShowRepository,
	bookingRepo BookingRepository,
	opdReadModel OpsReadModel,
) *echo.Echo {
	e := libHttp.NewEcho()

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	handler := Handler{
		commandBus:            commandBus,
		eventBus:              eventBus,
		spreadsheetsAPIClient: spreadsheetsAPIClient,
		ticketRepo:            ticketRepo,
		showRepo:              showRepo,
		bookingRepo:           bookingRepo,
		opsReadModel:          opdReadModel,
	}

	e.POST("/tickets-status", handler.PostTicketsStatus)
	e.GET("/tickets", handler.GetAllTickets)
	e.POST("/shows", handler.PostShow)
	e.POST("/book-tickets", handler.PostBookTickets)
	e.PUT("/ticket-refund/:ticket_id", handler.PutTicketRefund)
	e.GET("/ops/bookings", handler.GetOpsTickets)
	e.GET("/ops/bookings/:id", handler.GetOpsTicket)

	return e
}
