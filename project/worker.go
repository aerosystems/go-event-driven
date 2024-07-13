package main

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"os"
	"time"
)

type MessageType int

const (
	defaultDelay     = 100 * time.Millisecond
	defaultQueueSize = 100

	Receipt MessageType = iota
	Spreadsheet
)

type Message struct {
	Type     MessageType
	TicketId string
}

type Worker struct {
	queue chan Message
}

func NewWorker() *Worker {
	return &Worker{
		queue: make(chan Message, defaultQueueSize),
	}
}

func (w Worker) Enqueue(messages ...Message) {
	for _, msg := range messages {
		w.queue <- msg
	}
}

func (w Worker) Run() {
	clients, err := clients.NewClients(os.Getenv("GATEWAY_ADDR"), nil)
	if err != nil {
		panic(err)
	}

	receiptsClient := NewReceiptsClient(clients)
	spreadsheetsClient := NewSpreadsheetsClient(clients)

	ctx := context.Background()

	for {
		msg := <-w.queue
		switch msg.Type {
		case Receipt:
			if err := receiptsClient.IssueReceipt(ctx, msg.TicketId); err != nil {
				time.Sleep(defaultDelay)
				w.Enqueue(msg)
			}
		case Spreadsheet:
			if err := spreadsheetsClient.AppendRow(ctx, "tickets-to-print", []string{msg.TicketId}); err != nil {
				time.Sleep(defaultDelay)
				w.Enqueue(msg)
			}
		default:
			panic("unhandled default case")
		}
	}
}
