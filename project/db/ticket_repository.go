package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	"strconv"
	"tickets/entities"
)

type TicketRepo struct {
	db *sqlx.DB
}

func NewTicketRepository(db *sqlx.DB) TicketRepo {
	if db == nil {
		panic("missing db")
	}
	return TicketRepo{db: db}
}

type Ticket struct {
	ID            string  `db:"ticket_id"`
	PriceAmount   float64 `db:"price_amount"`
	PriceCurrency string  `db:"price_currency"`
	CustomerEmail string  `db:"customer_email"`
}

func entityToTicket(ticket entities.Ticket) (Ticket, error) {
	priceAmount, err := strconv.ParseFloat(ticket.Price.Amount, 64)
	if err != nil {
		return Ticket{}, err
	}
	return Ticket{
		ID:            ticket.TicketID,
		PriceAmount:   priceAmount,
		PriceCurrency: ticket.Price.Currency,
		CustomerEmail: ticket.CustomerEmail,
	}, nil
}

func (r TicketRepo) Add(ctx context.Context, t entities.Ticket) error {
	ticket, err := entityToTicket(t)
	if err != nil {
		return err
	}
	_, err = r.db.NamedExecContext(ctx, `
		INSERT INTO tickets (ticket_id, price_amount, price_currency, customer_email)
		VALUES (:ticket_id, :price_amount, :price_currency, :customer_email)
	`, ticket)
	return err
}
