package order_app

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
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

type Products []Product

func (a Products) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Products) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

type Order struct {
	Id         string   `json:"id,omitempty" db:"id"`
	Products   Products `json:"products" db:"products"`
	ShippingTo string   `json:"shipping_to" db:"shipping_to"`
	CardParams string   `json:"card_params" db:"card_params"`
	Status     string   `json:"status" db:"status"` // "CREATED"/"CANCELED"/"COMPLETED"/"PAYED"/"RESERVED"/"DELIVERED"
}

type OrderStatus struct {
	Time    time.Time `json:"time,omitempty" db:"time"`
	IdOrder string    `json:"id_order" db:"id_order"`
	Status  string    `json:"status" db:"status"`
}

type OrderID struct {
	Id string `json:"id,omitempty"`
}

type OrderEvent struct {
	Id     string `json:"id,omitempty"`
	Status string `json:"status"`
}

type Storage interface {
	CreateSchema() error
	CreateOrder(order Order) error
	UpdateOrderStatus(idOrder string, status string) error
	GetOrder(id string) (Order, error)
	GetListStatus(idOrder string) ([]OrderStatus, error)
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
	Warn(msg string)
}

type OrderMQ interface {
	PublishOrder(order Order) error
	PublishStatus(idOrder string, statusOrder string) error
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
	return a.mq.PublishOrder(order)
}

func (a *SrvOrder) CancelOrder(id string) error {
	return a.mq.PublishStatus(id, "CANCELED")
}

func (a *SrvOrder) StatusOrder(id string) (string, error) {
	order, err := a.storage.GetOrder(id)
	if err != nil {
		return "", err
	}
	return order.Status, nil
}

func (a *SrvOrder) StatusOrderChangeList(id string) ([]string, error) {
	statusList, err := a.storage.GetListStatus(id)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(statusList))
	for i := 0; i < len(statusList); i++ {
		result[i] = statusList[i].Status
	}
	return result, nil
}
