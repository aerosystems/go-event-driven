package main

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"net/http"
	"os"
	"os/signal"
	"tickets/api"
	"tickets/message"
	"tickets/service"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	traceHttpClient := &http.Client{Transport: otelhttp.NewTransport(
		http.DefaultTransport,
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("HTTP %s %s %s", r.Method, r.URL.String(), operation)
		}),
	)}

	apiClients, err := clients.NewClientsWithHttpClient(
		os.Getenv("GATEWAY_ADDR"),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
		traceHttpClient,
	)
	if err != nil {
		panic(err)
	}

	traceDB, err := otelsql.Open("postgres", os.Getenv("POSTGRES_URL"),
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithDBName("db"))
	if err != nil {
		panic(err)
	}

	db := sqlx.NewDb(traceDB, "postgres")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	redisClient := message.NewRedisClient(os.Getenv("REDIS_ADDR"))
	defer redisClient.Close()

	spreadsheetsService := api.NewSpreadsheetsAPIClient(apiClients)
	receiptsService := api.NewReceiptsServiceClient(apiClients)
	filesService := api.NewFilesServiceClient(apiClients)
	paymentsService := api.NewPaymentsServiceClient(apiClients)
	deadNationService := api.NewDeadNationServiceClient(apiClients)

	err = service.New(
		db,
		redisClient,
		spreadsheetsService,
		receiptsService,
		filesService,
		paymentsService,
		deadNationService,
	).Run(ctx)
	if err != nil {
		panic(err)
	}
}
