package orders_sqlstorage

import (
	"context"
	"encoding/json"
	"fmt"

	// Register pgx driver for postgresql.
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	orders_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw06_order/internal/orders/app"
)

type Storage struct {
	db *sqlx.DB
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Connect(ctx context.Context, dsn string) error {
	var err error
	s.db, err = sqlx.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("cannot open pgx driver: %w", err)
	}
	return s.db.PingContext(ctx)
}

func (s *Storage) CreateSchema() error {
	var err error
	_, err = s.db.Query(`CREATE TABLE IF NOT EXISTS orders (id text primary key, products jsonb, shipping_to text);`)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) CreateOrder(order orders_app.Order) error {
	if len(order.Id) == 0 {
		return orders_app.ErrOrderIDNotSet
	}
	productsStr, err := json.Marshal(order.Products)
	if err != nil {
		return err
	}

	_, err = s.db.NamedExec(`INSERT INTO orders (id, products, shipping_to)
		 VALUES (:id,:products,:shipping_to)`,
		map[string]interface{}{
			"id":          order.Id,
			"products":    string(productsStr),
			"shipping_to": order.ShippingTo,
		})
	return err
}
