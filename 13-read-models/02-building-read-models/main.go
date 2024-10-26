package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/shopspring/decimal"
)

type InvoiceIssued struct {
	InvoiceID    string
	CustomerName string
	Amount       decimal.Decimal
	IssuedAt     time.Time
}

type InvoicePaymentReceived struct {
	PaymentID  string
	InvoiceID  string
	PaidAmount decimal.Decimal
	PaidAt     time.Time

	FullyPaid bool
}

type InvoiceVoided struct {
	InvoiceID string
	VoidedAt  time.Time
}

type InvoiceReadModel struct {
	InvoiceID    string
	CustomerName string
	Amount       decimal.Decimal
	IssuedAt     time.Time

	FullyPaid     bool
	PaidAmount    decimal.Decimal
	LastPaymentAt time.Time

	Voided   bool
	VoidedAt time.Time
}

type InvoiceReadModelStorage struct {
	invoices map[string]InvoiceReadModel
}

func NewInvoiceReadModelStorage() *InvoiceReadModelStorage {
	return &InvoiceReadModelStorage{
		invoices: make(map[string]InvoiceReadModel),
	}
}

func (s *InvoiceReadModelStorage) Invoices() []InvoiceReadModel {
	invoices := make([]InvoiceReadModel, 0, len(s.invoices))
	for _, invoice := range s.invoices {
		invoices = append(invoices, invoice)
	}
	return invoices
}

func (s *InvoiceReadModelStorage) InvoiceByID(id string) (InvoiceReadModel, bool) {
	invoice, ok := s.invoices[id]
	return invoice, ok
}

func (s *InvoiceReadModelStorage) OnInvoiceIssued(_ context.Context, event *InvoiceIssued) error {
	s.invoices[event.InvoiceID] = InvoiceReadModel{
		InvoiceID:    event.InvoiceID,
		CustomerName: event.CustomerName,
		Amount:       event.Amount,
		IssuedAt:     event.IssuedAt,
	}
	return nil
}

func (s *InvoiceReadModelStorage) OnInvoicePaymentReceived(_ context.Context, event *InvoicePaymentReceived) error {
	invoice, ok := s.invoices[event.InvoiceID]
	if !ok {
		return fmt.Errorf("invoice %s not found", event.InvoiceID)
	}
	if invoice.FullyPaid {
		return nil
	}
	s.invoices[event.InvoiceID] = InvoiceReadModel{
		InvoiceID:     invoice.InvoiceID,
		CustomerName:  invoice.CustomerName,
		Amount:        invoice.Amount,
		IssuedAt:      invoice.IssuedAt,
		PaidAmount:    invoice.PaidAmount.Add(event.PaidAmount),
		LastPaymentAt: event.PaidAt,
		FullyPaid:     event.FullyPaid,
	}
	return nil
}

func (s *InvoiceReadModelStorage) OnInvoiceVoided(_ context.Context, event *InvoiceVoided) error {
	invoice, ok := s.invoices[event.InvoiceID]
	if !ok {
		return fmt.Errorf("invoice %s not found", event.InvoiceID)
	}
	s.invoices[event.InvoiceID] = InvoiceReadModel{
		InvoiceID:     invoice.InvoiceID,
		CustomerName:  invoice.CustomerName,
		Amount:        invoice.Amount,
		IssuedAt:      invoice.IssuedAt,
		PaidAmount:    invoice.PaidAmount,
		LastPaymentAt: invoice.LastPaymentAt,
		FullyPaid:     invoice.FullyPaid,
		Voided:        true,
		VoidedAt:      event.VoidedAt,
	}
	return nil
}

func NewRouter(storage *InvoiceReadModelStorage, eventProcessorConfig cqrs.EventProcessorConfig, watermillLogger watermill.LoggerAdapter) (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		return nil, fmt.Errorf("could not create router: %w", err)
	}

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, eventProcessorConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create command processor: %w", err)
	}

	err = eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"InvoiceIssued",
			storage.OnInvoiceIssued,
		),
		cqrs.NewEventHandler(
			"InvoicePaymentReceived",
			storage.OnInvoicePaymentReceived,
		),
		cqrs.NewEventHandler(
			"InvoiceVoided",
			storage.OnInvoiceVoided,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not add event handlers: %w", err)
	}

	return router, nil
}
