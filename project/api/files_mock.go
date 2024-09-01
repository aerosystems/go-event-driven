package api

import (
	"context"
	"fmt"
	"sync"
	"tickets/entities"
)

type FilesMock struct {
	mock sync.Mutex

	Files map[string]string
}

func (f *FilesMock) PrintTicket(ctx context.Context, ticket entities.Ticket) (string, error) {
	f.mock.Lock()
	defer f.mock.Unlock()

	fileID := fmt.Sprintf("%s-ticket.html", ticket.TicketID)
	fileBody := fmt.Sprintf("Ticket ID: %s. Price: %s %s", ticket.TicketID, ticket.Price.Amount, ticket.Price.Currency)

	if f.Files == nil {
		f.Files = make(map[string]string)
	}

	if _, ok := f.Files[fileID]; ok {
		return "", nil
	}

	f.Files[fileID] = fileBody
	return fileID, nil
}
