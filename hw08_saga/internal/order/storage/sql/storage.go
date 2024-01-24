package order_sqlstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	// Register pgx driver for postgresql.
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	order_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/order/app"
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
	_, err = s.db.Query(`CREATE TABLE IF NOT EXISTS orders (id text primary key, products jsonb,
		 shipping_to text, card_params text, status text);`)
	if err != nil {
		return err
	}
	_, err = s.db.Query(`CREATE TABLE IF NOT EXISTS history_status_change (time timestamptz primary key,
		id_order text, status text);`)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) CreateOrder(order order_app.Order) error {
	if len(order.Id) == 0 {
		return order_app.ErrOrderIDNotSet
	}
	productsStr, err := json.Marshal(order.Products)
	if err != nil {
		return err
	}
	_, err = s.db.NamedExec(`INSERT INTO orders (id, products, shipping_to, card_params, status)
		 VALUES (:id,:products,:shipping_to,:card_params,:status)`,
		map[string]interface{}{
			"id":          order.Id,
			"products":    string(productsStr),
			"shipping_to": order.ShippingTo,
			"card_params": order.CardParams,
			"status":      order.Status,
		})
	if err != nil {
		return err
	}
	_, err = s.db.NamedExec(`INSERT INTO history_status_change (time, id_order, status)
		 VALUES (:time,:id_order,:status)`,
		map[string]interface{}{
			"time":     time.Now(),
			"id_order": order.Id,
			"status":   order.Status,
		})
	return err
}

func (s *Storage) GetOrder(id string) (order_app.Order, error) {
	order := order_app.Order{}
	rows, err := s.db.NamedQuery(`SELECT id,products,shipping_to,card_params,status FROM orders WHERE id = :id`,
		map[string]interface{}{
			"id": id,
		})
	if err != nil {
		return order, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.StructScan(&order)
		if err != nil {
			return order, err
		}
	}
	return order, nil
}

func (s *Storage) UpdateOrderStatus(idOrder string, status string) error {
	result, err := s.db.NamedExec(`UPDATE orders SET status=:status 
	WHERE id = `+`'`+idOrder+`'`,
		map[string]interface{}{
			"status": status,
		})
	if result != nil {
		rowAffected, errResult := result.RowsAffected()
		if err == nil && rowAffected == 0 && errResult == nil {
			return order_app.ErrOrderIDNotExist
		}
	}
	_, err = s.db.NamedExec(`INSERT INTO history_status_change (time, id_order, status)
		 VALUES (:time,:id_order,:status)`,
		map[string]interface{}{
			"time":     time.Now(),
			"id_order": idOrder,
			"status":   status,
		})
	return err
}

func (s *Storage) GetListStatus(idOrder string) ([]order_app.OrderStatus, error) {
	orderStatus := make([]order_app.OrderStatus, 0)
	rows, err := s.db.NamedQuery(`SELECT time,id_order,status FROM history_status_change WHERE id_order = :id_order ORDER BY time`,
		map[string]interface{}{
			"id_order": idOrder,
		})
	if err != nil {
		return orderStatus, err
	}
	defer rows.Close()
	st := order_app.OrderStatus{}
	for rows.Next() {
		err := rows.StructScan(&st)
		if err != nil {
			return orderStatus, err
		}
		orderStatus = append(orderStatus, st)
	}
	return orderStatus, nil
}
