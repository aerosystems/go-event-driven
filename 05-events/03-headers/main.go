package main

import (
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Header struct {
	ID         string `json:"id"`
	EventName  string `json:"event_name"`
	OccurredAt string `json:"occurred_at"`
}

type ProductOutOfStock struct {
	Header    Header `json:"header"`
	ProductID string `json:"product_id"`
}

type ProductBackInStock struct {
	Header    Header `json:"header"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Publisher struct {
	pub message.Publisher
}

func NewPublisher(pub message.Publisher) Publisher {
	return Publisher{
		pub: pub,
	}
}

func NewProductOutOfStockMessage(productID string) *message.Message {
	event := ProductOutOfStock{
		Header: Header{
			ID:         watermill.NewUUID(),
			EventName:  "ProductOutOfStock",
			OccurredAt: time.Now().Format(time.RFC3339),
		},
		ProductID: productID,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return nil
	}

	return message.NewMessage(event.Header.ID, payload)
}

func (p Publisher) PublishProductOutOfStock(productID string) error {
	return p.pub.Publish("product-updates", NewProductOutOfStockMessage(productID))
}

func NewProductBackInStockMessage(productID string, quantity int) *message.Message {
	event := ProductBackInStock{
		Header: Header{
			ID:         watermill.NewUUID(),
			EventName:  "ProductBackInStock",
			OccurredAt: time.Now().Format(time.RFC3339),
		},
		ProductID: productID,
		Quantity:  quantity,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return nil
	}

	return message.NewMessage(event.Header.ID, payload)
}

func (p Publisher) PublishProductBackInStock(productID string, quantity int) error {
	return p.pub.Publish("product-updates", NewProductBackInStockMessage(productID, quantity))
}
