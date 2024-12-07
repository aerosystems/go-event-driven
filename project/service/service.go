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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	"tickets/migrations"
	"time"
)

var (
	veryImportantCounter = promauto.NewCounter(prometheus.CounterOpts{
		// metric will be named tickets_very_important_counter_total
		Namespace: "tickets",
		Name:      "very_important_counter_total",
		Help:      "Total number of very important things processed",
	})
)

func init() {
	log.Init(logrus.InfoLevel)
}

type Service struct {
	db              *sqlx.DB
	watermillRouter *watermillMessage.Router
	echoRouter      *echo.Echo

	dataLake     db.DataLakeRepository
	opsReadModel db.OpsBookingReadModel
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

	var redisSubscriber watermillMessage.Subscriber
	redisSubscriber = message.NewRedisSubscriber(redisClient, watermillLogger)

	commandBus := command.NewBus(redisPublisher, command.NewBusConfig(watermillLogger))
	eventBus := event.NewBus(redisPublisher)
	ticketsRepo := db.NewTicketRepository(dbConn)
	showsRepo := db.NewShowRepository(dbConn)
	bookingRepo := db.NewBookingRepository(dbConn)
	opsBookingRepo := db.NewOpsBookingReadModel(dbConn, eventBus)
	dataLakeRepo := db.NewDataLakeRepository(dbConn)

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
		eventBus,
		receiptsService,
		paymentsService,
	)

	postgresSubscriber := outbox.NewPostgresSubscriber(dbConn.DB, watermillLogger)

	eventProcessorConfig := event.NewProcessorConfig(redisClient, watermillLogger)
	commandProcessorConfig := command.NewProcessorConfig(redisClient, watermillLogger)

	watermillRouter := message.NewWatermillRouter(
		postgresSubscriber,
		redisPublisher,
		redisSubscriber,
		eventProcessorConfig,
		eventsHandler,
		commandProcessorConfig,
		commandHandler,
		opsBookingRepo,
		dataLakeRepo,
		watermillLogger,
	)
	echoRouter := ticketsHttp.NewHttpRouter(
		commandBus,
		eventBus,
		spreadsheetsService,
		ticketsRepo,
		showsRepo,
		bookingRepo,
		opsBookingRepo,
	)

	return Service{
		dbConn,
		watermillRouter,
		echoRouter,
		dataLakeRepo,
		opsBookingRepo,
	}
}

func (s Service) Run(
	ctx context.Context,
) error {
	if err := db.InitializeDatabaseSchema(s.db); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	go func() {
		if err := migrations.MigrateReadModel(ctx, s.dataLake, s.opsReadModel); err != nil {
			log.FromContext(ctx).Errorf("failed to migrate read model: %v", err)
		}
	}()

	go func() {
		for {
			veryImportantCounter.Inc()
			time.Sleep(time.Millisecond * 100)
		}
	}()

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
