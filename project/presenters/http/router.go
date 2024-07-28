package HttpRouter

import (
	"context"
	"errors"
	"fmt"
	commonHTTP "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	HttpTicketHandler "tickets/presenters/http/handlers/ticket"
)

const webPort = 8080

type Router struct {
	echo *echo.Echo
}

func NewRouter(log *logrus.Logger, ticketHandler *HttpTicketHandler.Handler) *Router {
	e := commonHTTP.NewEcho()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			log.WithFields(logrus.Fields{
				"URI":    values.URI,
				"status": values.Status,
			}).Info("request")

			return nil
		},
	}))
	e.Use(middleware.Recover())

	e.GET("/health", ticketHandler.Health)
	e.POST("/tickets-status", ticketHandler.TicketsStatus)

	return &Router{
		e,
	}
}

func (r Router) Run() error {
	err := r.echo.Start(fmt.Sprintf(":%d", webPort))
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (r Router) Shutdown(ctx context.Context) error {
	return r.echo.Shutdown(ctx)
}
