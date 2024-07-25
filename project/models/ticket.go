package models

type Ticket struct {
	TicketID      string      `json:"ticket_id"`
	Status        string      `json:"status"`
	CustomerEmail string      `json:"customer_email"`
	Price         TicketPrice `json:"price"`
}

type TicketPrice struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}
