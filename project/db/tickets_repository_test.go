package db

import (
	"context"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"os"
	"sync"
	"testing"
	"tickets/entities"
)

var (
	db     *sqlx.DB
	single sync.Once
)

func setupDB() *sqlx.DB {
	single.Do(func() {
		var err error
		db, err = sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
		if err != nil {
			panic(err)
		}
	})
	return db
}

func TestTicketRepository(t *testing.T) {
	db := setupDB()
	InitializeDatabaseSchema(db)
	ticketRepo := NewTicketRepository(db)

	ticket := entities.Ticket{
		TicketID: uuid.New().String(),
		Price: entities.Money{
			Amount:   "50",
			Currency: "USD",
		},
		CustomerEmail: "example@test.com",
	}

	countAttempts := 2
	for i := 0; i < countAttempts; i++ {
		err := ticketRepo.Add(context.Background(), ticket)
		require.NoError(t, err)
	}

	var tickets []Ticket
	err := db.SelectContext(
		context.Background(),
		&tickets,
		`SELECT * FROM tickets WHERE ticket_id = $1`,
		ticket.TicketID)
	require.NoError(t, err)
	require.Len(t, tickets, 1)
}
