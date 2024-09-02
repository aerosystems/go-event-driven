package entities

import "time"

type Show struct {
	ShowID          string
	DeadNationID    string
	NumberOfTickets int
	StartTime       time.Time
	Title           string
	Venue           string
}
