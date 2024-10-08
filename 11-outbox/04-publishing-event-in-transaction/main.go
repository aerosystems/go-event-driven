package main

import (
	"database/sql"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	_ "github.com/lib/pq"
)

func PublishInTx(
	message *message.Message,
	tx *sql.Tx,
	logger watermill.LoggerAdapter,
) error {
	publisher, err := watermillSQL.NewPublisher(tx, watermillSQL.PublisherConfig{
		SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
	}, logger)
	if err != nil {
		return err
	}
	publisher.Publish("ItemAddedToCart", message)
	return nil
}
