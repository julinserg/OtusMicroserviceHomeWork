package pay_sqlstorage

import (
	"context"
	"fmt"
	"time"

	// Register pgx driver for postgresql.
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	pay_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/pay/app"
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
	_, err = s.db.Query(`CREATE TABLE IF NOT EXISTS pay (id text primary key, id_order text, card_params text, 
		amount int, operation text, time timestamptz, CONSTRAINT fk_order FOREIGN KEY(id_order) REFERENCES orders(id));`)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) CreatePaymentOperation(payment pay_app.PayOperation) error {
	_, err := s.db.NamedExec(`INSERT INTO pay (id, id_order, card_params, amount, operation, time)
		 VALUES (:id,:id_order,:card_params,:amount,:operation, :time)`,
		map[string]interface{}{
			"id":          payment.Id,
			"id_order":    payment.IdOrder,
			"card_params": payment.CardParams,
			"amount":      payment.Amount,
			"operation":   payment.Operation,
			"time":        time.Now(),
		})
	return err
}

func (s *Storage) GetPaymentOperation(idOrder string) (pay_app.PayOperation, error) {
	operation := pay_app.PayOperation{}
	rows, err := s.db.NamedQuery(`SELECT id,id_order,card_params,amount,operation FROM pay WHERE id_order=:id_order`,
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
