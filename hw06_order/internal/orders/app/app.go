package orders_app

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
}

type Storage interface {
	CreateSchema() error
	CreateOrder(order Order) error
	GetRequest(id string) (Request, error)
	SaveRequest(obj Request) error
}
