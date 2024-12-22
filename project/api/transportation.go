package api

import (
	"context"
	"fmt"
	"net/http"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/transportation"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type TransportationClient struct {
	clients *clients.Clients
}

func NewTransportationClient(clients *clients.Clients) *TransportationClient {
	if clients == nil {
		panic("clients is nil")
	}

	return &TransportationClient{clients: clients}
}

func (t TransportationClient) BookFlight(
	ctx context.Context,
	request entities.BookFlightTicketRequest,
) (entities.BookFlightTicketResponse, error) {
	resp, err := t.clients.Transportation.PutFlightTicketsWithResponse(ctx, transportation.BookFlightTicketRequest{
		CustomerEmail:  request.CustomerEmail,
		FlightId:       request.FlightID,
		PassengerNames: request.PassengerNames,
		ReferenceId:    request.ReferenceId,
		IdempotencyKey: request.IdempotencyKey,
	})
	if err != nil {
		return entities.BookFlightTicketResponse{}, fmt.Errorf("failed to book flight: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		return entities.BookFlightTicketResponse{
			TicketIds: lo.Map(resp.JSON201.TicketIds, func(i openapi_types.UUID, _ int) uuid.UUID {
				return i
			}),
		}, nil
	case http.StatusConflict:
		return entities.BookFlightTicketResponse{}, entities.ErrNoFlightTicketsAvailable
	default:
		return entities.BookFlightTicketResponse{}, fmt.Errorf(
			"unexpected status code for PUT transportation-api/transportation/flight-tickets: %d",
			resp.StatusCode(),
		)
	}
}

func (t TransportationClient) BookTaxi(ctx context.Context, request entities.BookTaxiRequest) (entities.BookTaxiResponse, error) {
	resp, err := t.clients.Transportation.PutTaxiBookingWithResponse(ctx, transportation.TaxiBookingRequest{
		CustomerEmail:      request.CustomerEmail,
		NumberOfPassengers: request.NumberOfPassengers,
		PassengerName:      request.PassengerName,
		ReferenceId:        request.ReferenceId,
		IdempotencyKey:     request.IdempotencyKey,
	})
	if err != nil {
		return entities.BookTaxiResponse{}, fmt.Errorf("failed to book taxi: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		return entities.BookTaxiResponse{
			TaxiBookingId: resp.JSON201.BookingId,
		}, nil
	default:
		return entities.BookTaxiResponse{}, fmt.Errorf(
			"unexpected status code for PUT transportation-api/transportation/taxi-tickets: %d",
			resp.StatusCode(),
		)
	}
}
