package PubSubTicketHandler

import (
	"encoding/json"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"tickets/models"
)

func (h Handler) ReceiptConfirm(msg *message.Message) error {
	if msg.UUID == brokenMsgUuid {
		return nil
	}
	if msg.Metadata.Get("type") != "TicketBookingConfirmed" {
		return nil
	}
	var payload models.TicketBookingConfirmed
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}
	if payload.Price.Currency == "" {
		payload.Price.Currency = "USD"
	}
	reqCorrelationID := msg.Metadata.Get("correlation_id")
	ctx := log.ContextWithCorrelationID(msg.Context(), reqCorrelationID)
	msg.SetContext(ctx)
	if err := h.receiptsClient.IssueReceipt(msg.Context(), models.IssueReceiptRequest{
		TicketID: payload.TicketID,
		Price:    payload.Price,
	}); err != nil {
		return err
	}
	return nil
}
