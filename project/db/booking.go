package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	"tickets/entities"
)

type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) BookingRepository {
	if db == nil {
		panic("missing db")
	}
	return BookingRepository{db: db}
}

type Booking struct {
	BookingID       string `db:"booking_id"`
	ShowID          string `db:"show_id"`
	NumberOfTickets int    `db:"number_of_tickets"`
	CustomerEmail   string `db:"customer_email"`
}

func entityToBooking(booking entities.Booking) (Booking, error) {
	return Booking{
		BookingID:       booking.BookingID,
		ShowID:          booking.ShowID,
		NumberOfTickets: booking.NumberOfTickets,
		CustomerEmail:   booking.CustomerEmail,
	}, nil
}

func (r BookingRepository) Create(ctx context.Context, booking entities.Booking) (string, error) {
	b, err := entityToBooking(booking)
	if err != nil {
		return "", err
	}
	_, err = r.db.NamedExecContext(ctx, `
		INSERT INTO bookings (booking_id, show_id, number_of_tickets, customer_email)
		VALUES (:booking_id, :show_id, :number_of_tickets, :customer_email)
		ON CONFLICT (booking_id) DO NOTHING
	`, b)
	return booking.BookingID, err
}
