package orders_sqlstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
	if err != nil {
		return err
	}
	_, err = s.db.Query(`CREATE TABLE IF NOT EXISTS requests (id text primary key, response_code int, error_text text);`)
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

func (s *Storage) GetOrdersCount() (int, error) {
	result := 0
	rows, err := s.db.Query(`SELECT COUNT(*) FROM orders`)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&result)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (s *Storage) UpdateRequest(obj orders_app.Request) error {
	if len(obj.Id) == 0 {
		return orders_app.ErrRequestIDNotSet
	}

	result, err := s.db.NamedExec(`UPDATE requests SET response_code=:response_code, 
	error_text=:error_text WHERE id = `+`'`+obj.Id+`'`,
		map[string]interface{}{
			"response_code": obj.Code,
			"error_text":    obj.ErrorText,
		})
	if result != nil {
		rowAffected, errResult := result.RowsAffected()
		if err == nil && rowAffected == 0 && errResult == nil {
			return orders_app.ErrRequestIDNotSet
		}
	}
	return err
}

func (s *Storage) GetOrCreateRequest(id string) (orders_app.Request, error) {
	req := orders_app.Request{IsNew: false}
	if len(id) == 0 {
		return req, orders_app.ErrRequestIDNotSet
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return req, err
	}
	defer tx.Rollback()

	rows, err := tx.NamedQuery(`SELECT * FROM requests WHERE id=:id`, map[string]interface{}{"id": id})
	if err != nil {
		return req, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.StructScan(&req)
		if err != nil {
			return req, err
		}
	}
	if len(req.Id) == 0 {
		_, err = tx.NamedExec(`INSERT INTO requests (id, response_code, error_text)
		 VALUES (:id,:response_code,:error_text)`,
			map[string]interface{}{
				"id":            id,
				"response_code": http.StatusAccepted,
				"error_text":    "",
			})
		if err != nil {
			return req, err
		}
		req.IsNew = true
	}
	err = tx.Commit()
	if err != nil {
		return req, err
	}
	return req, nil
}
