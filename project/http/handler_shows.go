package http

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"tickets/entities"
	"time"
)

type ShowRequestBody struct {
	DeadNationID    string `json:"dead_nation_id"`
	NumberOfTickets int    `json:"number_of_tickets"`
	StartTime       string `json:"start_time"`
	Title           string `json:"title"`
	Venue           string `json:"venue"`
}

type ShowResponseBody struct {
	ShowID string `json:"show_id"`
}

func (h Handler) PostShow(c echo.Context) error {
	ctx := c.Request().Context()

	var req ShowRequestBody
	if err := c.Bind(&req); err != nil {
		return err
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return err
	}

	show := entities.Show{
		ShowID:          uuid.New().String(),
		DeadNationID:    req.DeadNationID,
		NumberOfTickets: req.NumberOfTickets,
		StartTime:       startTime,
		Title:           req.Title,
		Venue:           req.Venue,
	}

	showID, err := h.showRepo.Create(ctx, show)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, ShowResponseBody{ShowID: showID})
}
