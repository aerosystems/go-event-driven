package service

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"os/signal"
	HttpRouter "tickets/presenters/http"
	HttpTicketHandler "tickets/presenters/http/handlers/ticket"
	PubSubRouter "tickets/presenters/pubsub"
	PubSubTicketHandler "tickets/presenters/pubsub/handlers/ticket"
	"tickets/presenters/pubsub/subs"
)

const webPort = 8080

type Service struct {
	httpRouter   *HttpRouter.Router
	pubsubRouter *PubSubRouter.Router
}

func NewService(redisClient *redis.Client, spreadsheetsClient SpreadsheetsClient, receiptsClient ReceiptsClient) *Service {
	logger := log.NewWatermill(log.FromContext(context.Background()))

	ticketPub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: redisClient,
	}, log.NewWatermill(logrus.NewEntry(logrus.StandardLogger())))
	if err != nil {
		panic(err)
	}

	httpTicketHandler := HttpTicketHandler.NewHttpTicketHandler(ticketPub)

	httpRouter := HttpRouter.NewRouter(webPort, httpTicketHandler)

	pubsubTicketHandler := PubSubTicketHandler.NewTicketHandler(spreadsheetsClient, receiptsClient)

	receiptsSub := subs.NewReceiptsSub(logger, redisClient)
	spreadsheetsSub := subs.NewSpreadsheetsSub(logger, redisClient)

	pubsubRouter := PubSubRouter.NewPubSubRouter(logger, pubsubTicketHandler, spreadsheetsSub, receiptsSub)

	return &Service{
		httpRouter,
		pubsubRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.pubsubRouter.Run()
	})
	g.Go(func() error {
		<-s.pubsubRouter.Running()
		return s.httpRouter.Run()
	})
	g.Go(func() error {
		<-ctx.Done()
		return s.httpRouter.Shutdown(ctx)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
	return nil
}
