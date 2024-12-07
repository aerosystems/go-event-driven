package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"tickets/entities"
)

type DataLakeRepository struct {
	db *sqlx.DB
}

func NewDataLakeRepository(db *sqlx.DB) DataLakeRepository {
	if db == nil {
		panic("nil db")
	}

	return DataLakeRepository{db: db}
}

func (e DataLakeRepository) AddEvent(ctx context.Context, event entities.DataLakeEvent) error {
	_, err := e.db.ExecContext(
		ctx,
		`INSERT INTO events (event_id, published_at, event_name, event_payload) VALUES ($1, $2, $3, $4)`,
		event.EventID,
		event.PublishedAt,
		event.EventName,
		event.EventPayload,
	)
	var postgresError *pq.Error
	if errors.As(err, &postgresError) && postgresError.Code.Name() == "unique_violation" {
		// handling re-delivery
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not store %s event in data lake: %w", event.EventID, err)
	}

	return nil
}

func (e DataLakeRepository) GetEvents(ctx context.Context) ([]entities.DataLakeEvent, error) {
	var events []entities.DataLakeEvent
	err := e.db.SelectContext(ctx, &events, "SELECT * FROM events ORDER BY published_at ASC")
	if err != nil {
		return nil, fmt.Errorf("could not get events from data lake: %w", err)
	}

	return events, nil
}
