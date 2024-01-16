package order_app

import (
	"errors"
)

var (
	ErrOrderIDNotSet   = errors.New("Order ID not set")
	ErrOrderIDNotExist = errors.New("Order ID not exist")
	ErrRequestIDNotSet = errors.New("Request ID not set")
)

type Request struct {
	Id        string `db:"id"`
	Code      int    `db:"response_code"`
	ErrorText string `db:"error_text"`
	IsNew     bool
}

type Product struct {
	Id    int64  `json:"id" db:"id"`
	Name  string `json:"name,omitempty" db:"name"`
	Price int    `json:"price" db:"price"`
}

type Order struct {
	Id         string    `json:"id,omitempty"`
	Products   []Product `json:"products"`
	ShippingTo string    `json:"shipping_to"`
	CardParams string    `json:"card_params"`
	Status     string    `json:"status"` // "CREATED"/"CANCELED"/"COMPLETED"/"PAYED"/"RESERVED"/"DELIVERED"
}

type OrderEvent struct {
	Id     string `json:"id,omitempty"`
	Status string `json:"status"`
}

type Storage interface {
	CreateSchema() error
	CreateOrder(order Order) error
	UpdateOrderStatus(idOrder string, status string) error
	GetOrdersCount() (int, error)
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
	Warn(msg string)
}

type OrderMQ interface {
	Publish(order Order) error
}

type SrvOrder struct {
	logger  Logger
	storage Storage
	mq      OrderMQ
}

func New(logger Logger, storage Storage, mq OrderMQ) *SrvOrder {
	return &SrvOrder{logger, storage, mq}
}

func (a *SrvOrder) CreateOrder(order Order) error {
	order.Status = "CREATED"
	err := a.storage.CreateOrder(order)
	if err != nil {
		return err
	}
	return a.mq.Publish(order)
}

func (a *SrvOrder) GetOrdersCount() (int, error) {
	return a.storage.GetOrdersCount()
}
