package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	"tickets/entities"
	"time"
)

type ShowRepository struct {
	db *sqlx.DB
}

func NewShowRepository(db *sqlx.DB) ShowRepository {
	if db == nil {
		panic("missing db")
	}
	return ShowRepository{db: db}
}

type Show struct {
	ShowID          string    `db:"show_id"`
	DeadNationID    string    `db:"dead_nation_id"`
	NumberOfTickets int       `db:"number_of_tickets"`
	StartTime       time.Time `db:"start_time"`
	Title           string    `db:"title"`
	Venue           string    `db:"venue"`
}

func entityToShow(show entities.Show) (Show, error) {
	return Show{
		ShowID:          show.ShowID,
		DeadNationID:    show.DeadNationID,
		NumberOfTickets: show.NumberOfTickets,
		StartTime:       show.StartTime,
		Title:           show.Title,
		Venue:           show.Venue,
	}, nil
}

func (r ShowRepository) Create(ctx context.Context, show entities.Show) (string, error) {
	s, err := entityToShow(show)
	if err != nil {
		return "", err
	}
	_, err = r.db.NamedExecContext(ctx, `
		INSERT INTO shows (show_id, dead_nation_id, number_of_tickets, start_time, title, venue)
		VALUES (:show_id, :dead_nation_id, :number_of_tickets, :start_time, :title, :venue)
		ON CONFLICT (show_id) DO NOTHING
	`, s)
	return show.ShowID, err
}

func (r ShowRepository) Get(ctx context.Context, showID string) (entities.Show, error) {
	var s Show
	err := r.db.GetContext(ctx, &s, "SELECT * FROM shows WHERE show_id = $1", showID)
	if err != nil {
		return entities.Show{}, err
	}
	return entities.Show{
		ShowID:          s.ShowID,
		DeadNationID:    s.DeadNationID,
		NumberOfTickets: s.NumberOfTickets,
		StartTime:       s.StartTime,
		Title:           s.Title,
		Venue:           s.Venue,
	}, nil
}
