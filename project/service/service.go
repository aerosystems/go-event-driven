package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	stdHTTP "net/http"
	"tickets/db"
	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/message/command"
	"tickets/message/event"
	"tickets/message/outbox"
)

const ForwarderTopic = "events_to_forward"

func init() {
	log.Init(logrus.InfoLevel)
}

type Service struct {
	db              *sqlx.DB
	redisClient     *redis.Client
	watermillRouter *watermillMessage.Router
	echoRouter      *echo.Echo
}

func New(
	dbConn *sqlx.DB,
	redisClient *redis.Client,
	spreadsheetsService event.SpreadsheetsAPI,
	receiptsService event.ReceiptsService,
	filesService event.FilesService,
	paymentsService command.PaymentsService,
	deadNationService event.DeadNationService,
) Service {
	watermillLogger := log.NewWatermill(log.FromContext(context.Background()))

	var redisPublisher watermillMessage.Publisher
	redisPublisher = message.NewRedisPublisher(redisClient, watermillLogger)
	redisPublisher = log.CorrelationPublisherDecorator{Publisher: redisPublisher}

	commandBus := command.NewBus(redisPublisher)
	eventBus := event.NewBus(redisPublisher)
	ticketsRepo := db.NewTicketRepository(dbConn)
	showsRepo := db.NewShowRepository(dbConn)
	bookingRepo := db.NewBookingRepository(dbConn)
	opsBookingRepo := db.NewOpsBookingReadModel(dbConn)

	eventsHandler := event.NewHandler(
		eventBus,
		spreadsheetsService,
		receiptsService,
		filesService,
		ticketsRepo,
		showsRepo,
		deadNationService,
	)

	commandHandler := command.NewHandler(
		commandBus,
		paymentsService,
	)

	postgresSubscriber := outbox.NewPostgresSubscriber(dbConn.DB, watermillLogger)

	eventProcessorConfig := event.NewProcessorConfig(redisClient, watermillLogger)
	commandProcessorConfig := command.NewProcessorConfig(redisClient, watermillLogger)

	watermillRouter := message.NewWatermillRouter(
		postgresSubscriber,
		redisPublisher,
		eventProcessorConfig,
		eventsHandler,
		commandProcessorConfig,
		commandHandler,
		opsBookingRepo,
		watermillLogger,
	)
	echoRouter := ticketsHttp.NewHttpRouter(
		commandBus,
		eventBus,
		spreadsheetsService,
		ticketsRepo,
		showsRepo,
		bookingRepo,
	)

	return Service{
		dbConn,
		redisClient,
		watermillRouter,
		echoRouter,
	}
}

func (s Service) Run(
	ctx context.Context,
) error {
	if err := db.InitializeDatabaseSchema(s.db); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.watermillRouter.Run(ctx)
	})

	g.Go(func() error {
		// we don't want to start HTTP server before Watermill router (so service won't be healthy before it's ready)
		<-s.watermillRouter.Running()

		err := s.echoRouter.Start(":8080")

		if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
			return err
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		return s.echoRouter.Shutdown(context.Background())
	})

	return g.Wait()
}
