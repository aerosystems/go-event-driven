package HttpTicketHandler

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h Handler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
