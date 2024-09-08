package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tickets/entities"
	"tickets/message/event"
	"tickets/message/outbox"

	"github.com/jmoiron/sqlx"
)

var ErrNotEnoughTickets = errors.New("not enough tickets available")

type BookingsRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) BookingsRepository {
	if db == nil {
		panic("nil db")
	}

	return BookingsRepository{db: db}
}

func (b BookingsRepository) AddBooking(ctx context.Context, booking entities.Booking) (err error) {
	opts := sql.TxOptions{Isolation: sql.LevelSerializable}
	tx, err := b.db.BeginTxx(ctx, &opts)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			err = errors.Join(err, rollbackErr)
			return
		}
		err = tx.Commit()
	}()

	var count int
	if err := tx.QueryRow(`
		SELECT number_of_tickets
		FROM shows
		WHERE show_id = $1
		FOR UPDATE
		`, booking.ShowID).Scan(&count); err != nil {
		return fmt.Errorf("could not get number of tickets: %w", err)
	}

	if count < booking.NumberOfTickets {
		return fmt.Errorf("%w: %d tickets available at this monent", ErrNotEnoughTickets, count)
	}

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO 
		    bookings (booking_id, show_id, number_of_tickets, customer_email) 
		VALUES (:booking_id, :show_id, :number_of_tickets, :customer_email)
		`, booking)
	if err != nil {
		return fmt.Errorf("could not add booking: %w", err)
	}

	outboxPublisher, err := outbox.NewPublisherForDb(ctx, tx)
	if err != nil {
		return fmt.Errorf("could not create event bus: %w", err)
	}

	err = event.NewBus(outboxPublisher).Publish(ctx, entities.BookingMade{
		Header:          entities.NewEventHeader(),
		BookingID:       booking.BookingID,
		NumberOfTickets: booking.NumberOfTickets,
		CustomerEmail:   booking.CustomerEmail,
		ShowId:          booking.ShowID,
	})
	if err != nil {
		return fmt.Errorf("could not publish event: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE shows
		SET number_of_tickets = number_of_tickets - $1
		WHERE show_id = $2
		`, booking.NumberOfTickets, booking.ShowID)
	if err != nil {
		return fmt.Errorf("could not update number of tickets: %w", err)
	}
	return nil
}
