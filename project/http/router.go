package http

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func NewHttpRouter(
	eventBus *cqrs.EventBus,
	commandBus *cqrs.CommandBus,
	spreadsheetsAPIClient SpreadsheetsAPI,
	ticketsRepository TicketsRepository,
	opsBookingReadModel OpsBookingReadModel,
	showsRepository ShowsRepository,
	bookingsRepository BookingsRepository,
	vipBundlesRepository VipBundlesRepository,
) *echo.Echo {
	e := libHttp.NewEcho()

	e.Use(otelecho.Middleware("tickets"))

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	handler := Handler{
		eventBus:              eventBus,
		commandBus:            commandBus,
		spreadsheetsAPIClient: spreadsheetsAPIClient,
		ticketsRepo:           ticketsRepository,
		opsBookingReadModel:   opsBookingReadModel,
		showsRepository:       showsRepository,
		bookingsRepository:    bookingsRepository,
		vipBundlesRepository:  vipBundlesRepository,
	}

	e.POST("/tickets-status", handler.PostTicketsStatus)

	e.PUT("/ticket-refund/:ticket_id", handler.PutTicketRefund)
	e.GET("/tickets", handler.GetTickets)
	e.POST("/book-tickets", handler.PostBookTickets)

	e.POST("/book-vip-bundle", handler.PostVipBundle)

	e.POST("/shows", handler.PostShows)

	e.GET("/ops/bookings", handler.GetOpsTickets)
	e.GET("/ops/bookings/:id", handler.GetOpsTicket)

	return e
}
