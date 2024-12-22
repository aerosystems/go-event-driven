package command

import (
	"golang.org/x/net/context"
	"tickets/entities"
)

func (h Handler) CancelFlightTickets(ctx context.Context, command *entities.CancelFlightTickets) error {
	return h.transportationServiceClient.CancelFlightTickets(
		ctx,
		entities.CancelFlightTicketsRequest{
			TicketIds: command.FlightTicketIDs,
		},
	)
}
