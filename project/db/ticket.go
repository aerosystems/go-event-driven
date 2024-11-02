package db

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"tickets/entities"
)

type TicketRepository struct {
	db *sqlx.DB
}

func NewTicketRepository(db *sqlx.DB) TicketRepository {
	if db == nil {
		panic("missing db")
	}
	return TicketRepository{db: db}
}

type Ticket struct {
	ID            string  `db:"ticket_id"`
	PriceAmount   float64 `db:"price_amount"`
	PriceCurrency string  `db:"price_currency"`
	CustomerEmail string  `db:"customer_email"`
	DeletedAt     *string `db:"deleted_at"`
}

func (r TicketRepository) Add(ctx context.Context, ticket entities.Ticket) error {
	_, err := r.db.NamedExecContext(
		ctx,
		`
		INSERT INTO 
    		tickets (ticket_id, price_amount, price_currency, customer_email) 
		VALUES 
		    (:ticket_id, :price.amount, :price.currency, :customer_email) 
		ON CONFLICT DO NOTHING`,
		ticket,
	)
	if err != nil {
		return fmt.Errorf("could not save ticket: %w", err)
	}

	return nil
}

func (r TicketRepository) Remove(ctx context.Context, ticketID string) error {
	res, err := r.db.ExecContext(
		ctx,
		`UPDATE tickets SET deleted_at = now() WHERE ticket_id = $1`,
		ticketID,
	)
	if err != nil {
		return fmt.Errorf("could not remove ticket: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("ticket with id %s not found", ticketID)
	}

	return nil
}

func (r TicketRepository) GetAll(ctx context.Context) ([]entities.Ticket, error) {
	var returnTickets []entities.Ticket

	err := r.db.SelectContext(
		ctx,
		&returnTickets, `
			SELECT 
				ticket_id,
				price_amount as "price.amount", 
				price_currency as "price.currency",
				customer_email
			FROM 
			    tickets
			WHERE
				deleted_at IS NULL
		`,
	)
	if err != nil {
		return nil, err
	}

	return returnTickets, nil
}
