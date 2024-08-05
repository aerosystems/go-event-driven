package api

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"net/http"
	"tickets/entities"
)

type FilesServiceClient struct {
	clients *clients.Clients
}

func NewFilesServiceClient(clients *clients.Clients) *FilesServiceClient {
	if clients == nil {
		panic("NewFilesServiceClient: clients is nil")
	}
	return &FilesServiceClient{
		clients: clients,
	}
}

func (f FilesServiceClient) PrintTicket(ctx context.Context, ticket entities.Ticket) error {
	fileID := fmt.Sprintf("%s-ticket.html", ticket.TicketID)
	fileBody := fmt.Sprintf("Ticket ID: %s. Price: %s %s", ticket.TicketID, ticket.Price.Amount, ticket.Price.Currency)
	response, err := f.clients.Files.PutFilesFileIdContentWithTextBodyWithResponse(ctx, fileID, fileBody)
	if err != nil {
		return err
	}
	if response.StatusCode() == http.StatusConflict {
		log.FromContext(ctx).Infof("file %s already exists", fileID)
		return nil
	}
	return nil
}
