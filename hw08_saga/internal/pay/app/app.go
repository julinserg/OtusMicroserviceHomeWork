package pay_app

import (
	"errors"

	"github.com/google/uuid"
	order_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/order/app"
)

var (
	ErrPayedAmountZero = errors.New("Payed error: amount is zero or negative")
)

type PayOperation struct {
	Id         string
	IdOrder    string
	CardParams string
	Amount     int
	Operation  string // "DEBIT"/"CREDIT"
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
	Warn(msg string)
}

type Storage interface {
	CreateSchema() error
	CreatePaymentOperation(payment PayOperation) error
}

type SrvPay struct {
	logger  Logger
	storage Storage
}

func New(logger Logger, storage Storage) *SrvPay {
	return &SrvPay{logger, storage}
}

func (s *SrvPay) CreatePaymentOperation(order order_app.Order) error {
	sum := 0
	for _, product := range order.Products {
		sum += product.Price
	}
	if sum <= 0 {
		return ErrPayedAmountZero
	}
	idPaymentOperation := uuid.New().String()
	paymentOperation := PayOperation{
		Id:         idPaymentOperation,
		IdOrder:    order.Id,
		CardParams: order.CardParams,
		Amount:     sum,
		Operation:  "DEBIT"}
	return s.storage.CreatePaymentOperation(paymentOperation)
}
