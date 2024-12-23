package orders

import (
	"github.com/jmoiron/sqlx"
)

func initializeDatabaseSchema(db *sqlx.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			order_id UUID PRIMARY KEY,
			shipped BOOLEAN NOT NULL,
			cancelled BOOLEAN NOT NULL
		);

		CREATE TABLE IF NOT EXISTS order_products (
			order_id UUID NOT NULL,
			product_id UUID NOT NULL,
			quantity INT NOT NULL,

		    PRIMARY KEY (order_id, product_id)
		);

		CREATE TABLE IF NOT EXISTS stock (
			product_id UUID PRIMARY KEY,
			quantity INT NOT NULL
		);
	`)
	if err != nil {
		panic(err)
	}
}

type ProductStock struct {
	ProductID string `db:"product_id" json:"product_id"`
	Quantity  int    `db:"quantity" json:"quantity"`
}
