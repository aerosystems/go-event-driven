package HttpRouter

import (
	"context"
	"errors"
	"fmt"
	commonHTTP "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/labstack/echo/v4"
	"net/http"
	HttpTicketHandler "tickets/presenters/http/handlers/ticket"
)

type Router struct {
	port          int
	ticketHandler *HttpTicketHandler.Handler
	echo          *echo.Echo
}

func NewRouter(port int, ticketHandler *HttpTicketHandler.Handler) *Router {
	return &Router{
		port,
		ticketHandler,
		commonHTTP.NewEcho(),
	}
}

func (r Router) Run() error {

	r.echo.GET("/health", r.ticketHandler.Health)
	r.echo.POST("/tickets-status", r.ticketHandler.TicketsStatus)

	err := r.echo.Start(fmt.Sprintf(":%d", r.port))
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (r Router) Shutdown(ctx context.Context) error {
	return r.echo.Shutdown(ctx)
}
