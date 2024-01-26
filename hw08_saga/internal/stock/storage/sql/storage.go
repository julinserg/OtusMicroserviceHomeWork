package stock_sqlstorage

import (
	"context"
	"fmt"
	"time"

	// Register pgx driver for postgresql.
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	stock_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/stock/app"
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
	_, err = s.db.Query(`CREATE TABLE IF NOT EXISTS stock (id text primary key, id_order text, id_product int, 
		count int, operation text, time timestamptz);`)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) CreateStockOperation(stock stock_app.StockOperation) error {
	_, err := s.db.NamedExec(`INSERT INTO stock (id, id_order, id_product, count, operation, time)
		 VALUES (:id,:id_order,:id_product,:count,:operation, :time)`,
		map[string]interface{}{
			"id":         stock.Id,
			"id_order":   stock.IdOrder,
			"id_product": stock.IdProduct,
			"count":      stock.Count,
			"operation":  stock.Operation,
			"time":       time.Now(),
		})
	return err
}

func (s *Storage) GetStockOperations(idOrder string) ([]stock_app.StockOperation, error) {
	operations := make([]stock_app.StockOperation, 0)
	rows, err := s.db.NamedQuery(`SELECT id,id_order,id_product,count,operation FROM stock WHERE id_order=:id_order`,
		map[string]interface{}{
			"id_order": idOrder,
		})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	operation := stock_app.StockOperation{}
	for rows.Next() {
		err := rows.StructScan(&operation)
		if err != nil {
			return nil, err
		}
		operations = append(operations, operation)
	}
	return operations, nil
}
