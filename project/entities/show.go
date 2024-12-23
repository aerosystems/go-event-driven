package entities

import (
	"github.com/google/uuid"
	"time"
)

type Show struct {
	ShowId          uuid.UUID `json:"show_id" db:"show_id"`
	DeadNationId    uuid.UUID `json:"dead_nation_id" db:"dead_nation_id"`
	NumberOfTickets int       `json:"number_of_tickets" db:"number_of_tickets"`
	StartTime       time.Time `json:"start_time" db:"start_time"`
	Title           string    `json:"title" db:"title"`
	Venue           string    `json:"venue" db:"venue"`
}
