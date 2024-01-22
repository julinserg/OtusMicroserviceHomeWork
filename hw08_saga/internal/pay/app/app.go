package pay_app

import (
	"errors"

	"github.com/google/uuid"
	order_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/order/app"
)

var (
	ErrPayedAmountZero      = errors.New("Payed error: amount is zero or negative")
	ErrPayedIdOrderNotFound = errors.New("Payed error: id order not found")
)

type PayOperation struct {
	Id         string `db:"id"`
	IdOrder    string `db:"id_order"`
	CardParams string `db:"card_params"`
	Amount     int    `db:"amount"`
	Operation  string `db:"operation"` // "DEBIT"/"CREDIT"
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
	GetPaymentOperation(idOrder string) (PayOperation, error)
}

type PayMQ interface {
	PublishOrder(order order_app.Order) error
	PublishStatus(idOrder string, statusOrder string) error
}

type SrvPay struct {
	logger  Logger
	storage Storage
	mq      PayMQ
}

func New(logger Logger, storage Storage, mq PayMQ) *SrvPay {
	return &SrvPay{logger, storage, mq}
}

func (s *SrvPay) createPaymentOperationLocal(order order_app.Order) error {
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

func (s *SrvPay) CreatePaymentOperation(order order_app.Order) error {

	err := s.createPaymentOperationLocal(order)
	if err == nil {
		s.mq.PublishStatus(order.Id, "PAYED")
		s.mq.PublishOrder(order)
	} else {
		s.mq.PublishStatus(order.Id, "CANCELED")
	}
	return err
}

func (s *SrvPay) RevertPaymentOperation(idOrder string, statusOrder string) error {
	if statusOrder != "CANCELED" {
		return nil
	}
	operation, err := s.storage.GetPaymentOperation(idOrder)
	if err != nil {
		return err
	}
	if len(operation.Id) == 0 {
		return ErrPayedIdOrderNotFound
	}
	operation.Id = uuid.New().String()
	operation.Operation = "CREDIT"
	err = s.storage.CreatePaymentOperation(operation)
	if err != nil {
		return err
	}
	return nil
}
