package stock_app

import (
	"errors"

	"github.com/google/uuid"
	order_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/order/app"
)

var (
	ErrStockCountZero       = errors.New("Stock error: count is zero or negative")
	ErrStockIdOrderNotFound = errors.New("Stock error: id order not found")
)

type StockOperation struct {
	Id        string `db:"id"`
	IdOrder   string `db:"id_order"`
	IdProduct int64  `db:"id_product"`
	Count     int    `db:"count"`
	Operation string `db:"operation"` // "RESERVED"/"UNRESERVED"
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
	Warn(msg string)
}

type Storage interface {
	CreateSchema() error
	CreateStockOperation(stock StockOperation) error
	GetStockOperations(idOrder string) ([]StockOperation, error)
}

type StockMQ interface {
	PublishOrder(order order_app.Order) error
	PublishStatus(idOrder string, statusOrder string) error
}

type SrvStock struct {
	logger  Logger
	storage Storage
	mq      StockMQ
}

func New(logger Logger, storage Storage, mq StockMQ) *SrvStock {
	return &SrvStock{logger, storage, mq}
}

func (s *SrvStock) createStockOperationLocal(order order_app.Order) error {
	for _, product := range order.Products {
		if product.Count <= 0 {
			return ErrStockCountZero
		}
		idStockOperation := uuid.New().String()
		stockOperation := StockOperation{
			Id:        idStockOperation,
			IdOrder:   order.Id,
			IdProduct: product.Id,
			Count:     product.Count,
			Operation: "RESERVED"}
		err := s.storage.CreateStockOperation(stockOperation)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SrvStock) CreateStockOperation(order order_app.Order) error {

	err := s.createStockOperationLocal(order)
	if err == nil {
		s.mq.PublishStatus(order.Id, "RESERVED")
		s.mq.PublishOrder(order)
	} else {
		s.mq.PublishStatus(order.Id, "CANCELED")
	}
	return err
}

func (s *SrvStock) RevertStockOperation(idOrder string, statusOrder string) error {
	if statusOrder != "CANCELED" {
		return nil
	}
	operations, err := s.storage.GetStockOperations(idOrder)
	if err != nil {
		return err
	}
	if len(operations) == 0 {
		return ErrStockIdOrderNotFound
	}
	for _, operation := range operations {
		operation.Id = uuid.New().String()
		operation.Operation = "UNRESERVED"
		err = s.storage.CreateStockOperation(operation)
		if err != nil {
			return err
		}
	}
	return nil
}
