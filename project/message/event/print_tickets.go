package event

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"tickets/entities"
)

func (h Handler) PrintTicket(ctx context.Context, event *entities.TicketBookingConfirmed_v1) error {
	log.FromContext(ctx).Info("Printing ticket")

	ticketHTML := `
		<html>
			<head>
				<title>Ticket</title>
			</head>
			<body>
				<h1>Ticket ` + event.TicketID + `</h1>
				<p>Price: ` + event.Price.Amount + ` ` + event.Price.Currency + `</p>	
			</body>
		</html>
`

	ticketFile := event.TicketID + "-ticket.html"

	err := h.filesAPI.UploadFile(ctx, ticketFile, ticketHTML)
	if err != nil {
		return fmt.Errorf("failed to upload ticket file: %w", err)
	}

	err = h.eventBus.Publish(ctx, entities.TicketPrinted_v1{
		Header:   entities.NewEventHeader(),
		TicketID: event.TicketID,
		FileName: ticketFile,
	})
	if err != nil {
		return fmt.Errorf("failed to publish TicketPrinted event: %w", err)
	}

	return nil
}
