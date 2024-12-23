package http

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"tickets/entities"
)

func (h Handler) PostShows(c echo.Context) error {
	show := entities.Show{}

	if err := c.Bind(&show); err != nil {
		return err
	}

	show.ShowId = uuid.New()

	if err := h.showsRepository.AddShow(c.Request().Context(), show); err != nil {
		return fmt.Errorf("failed to add show: %w", err)
	}

	return c.JSON(http.StatusCreated, show)
}
