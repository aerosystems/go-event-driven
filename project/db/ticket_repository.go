package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	"strconv"
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

func (r TicketRepository) Add(ctx context.Context, t entities.Ticket) error {
	ticket, err := entityToTicket(t)
	if err != nil {
		return err
	}
	_, err = r.db.NamedExecContext(ctx, `
		INSERT INTO tickets (ticket_id, price_amount, price_currency, customer_email)
		VALUES (:ticket_id, :price_amount, :price_currency, :customer_email)
		ON CONFLICT (ticket_id) DO NOTHING
	`, ticket)
	return err
}

func (r TicketRepository) Remove(ctx context.Context, ticketID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM tickets WHERE ticket_id = $1
	`, ticketID)
	return err
}

func (r TicketRepository) GetAll(ctx context.Context) ([]entities.Ticket, error) {
	var tickets []Ticket
	err := r.db.SelectContext(ctx, &tickets, "SELECT * FROM tickets")
	if err != nil {
		return nil, err
	}

	result := make([]entities.Ticket, 0, len(tickets))
	for _, ticket := range tickets {
		result = append(result, entities.Ticket{
			TicketID: ticket.ID,
			Price: entities.Money{
				Amount:   strconv.FormatFloat(ticket.PriceAmount, 'f', 2, 64),
				Currency: ticket.PriceCurrency,
			},
			CustomerEmail: ticket.CustomerEmail,
		})
	}
	return result, nil
}
