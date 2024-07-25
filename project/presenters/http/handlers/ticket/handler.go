package HttpTicketHandler

import "github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"

type Handler struct {
	ticketPub *redisstream.Publisher
}

func NewHttpTicketHandler(ticketPub *redisstream.Publisher) *Handler {
	return &Handler{
		ticketPub,
	}
}
