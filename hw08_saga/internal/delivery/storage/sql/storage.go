package delivery_sqlstorage

import (
	"context"
	"fmt"
	"time"

	// Register pgx driver for postgresql.
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	delivery_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/delivery/app"
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
	_, err = s.db.Query(`CREATE TABLE IF NOT EXISTS delivery (id text primary key, id_order text, shipping_to text, 
		operation text, time timestamptz);`)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) CreateDeliveryOperation(delivery delivery_app.DeliveryOperation) error {
	_, err := s.db.NamedExec(`INSERT INTO delivery (id, id_order, shipping_to, operation, time)
		 VALUES (:id,:id_order,:shipping_to,:operation, :time)`,
		map[string]interface{}{
			"id":          delivery.Id,
			"id_order":    delivery.IdOrder,
			"shipping_to": delivery.ShippingTo,
			"operation":   delivery.Operation,
			"time":        time.Now(),
		})
	return err
}

func (s *Storage) GetDeliveryOperation(idOrder string) (delivery_app.DeliveryOperation, error) {
	operation := delivery_app.DeliveryOperation{}
	rows, err := s.db.NamedQuery(`SELECT id,id_order,shipping_to,operation FROM delivery WHERE id_order=:id_order`,
		map[string]interface{}{
			"id_order": idOrder,
		})
	if err != nil {
		return operation, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.StructScan(&operation)
		if err != nil {
			return operation, err
		}
	}
	return operation, nil
}
