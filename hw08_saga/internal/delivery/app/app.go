package delivery_app

import (
	"errors"

	"github.com/google/uuid"
	order_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/order/app"
)

var (
	ErrDeliveryEmptyAddress    = errors.New("Delivery error: address is empty")
	ErrDeliveryIdOrderNotFound = errors.New("Delivery error: id order not found")
)

type DeliveryOperation struct {
	Id         string `db:"id"`
	IdOrder    string `db:"id_order"`
	ShippingTo string `db:"shipping_to"`
	Operation  string `db:"operation"` // "DELIVERED"/"CANCELED"
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
	Warn(msg string)
}

type Storage interface {
	CreateSchema() error
	CreateDeliveryOperation(delivery DeliveryOperation) error
	GetDeliveryOperation(idOrder string) (DeliveryOperation, error)
}

type DeliveryMQ interface {
	PublishStatus(idOrder string, statusOrder string) error
}

type SrvDelivery struct {
	logger  Logger
	storage Storage
	mq      DeliveryMQ
}

func New(logger Logger, storage Storage, mq DeliveryMQ) *SrvDelivery {
	return &SrvDelivery{logger, storage, mq}
}

func (s *SrvDelivery) createDeliveryOperationLocal(order order_app.Order) error {
	if len(order.ShippingTo) == 0 {
		return ErrDeliveryEmptyAddress
	}
	idDeliveryOperation := uuid.New().String()
	deliveryOperation := DeliveryOperation{
		Id:         idDeliveryOperation,
		IdOrder:    order.Id,
		ShippingTo: order.ShippingTo,
		Operation:  "DELIVERED"}
	err := s.storage.CreateDeliveryOperation(deliveryOperation)
	if err != nil {
		return err
	}
	return nil
}

func (s *SrvDelivery) CreateDeliveryOperation(order order_app.Order) error {

	err := s.createDeliveryOperationLocal(order)
	if err == nil {
		s.mq.PublishStatus(order.Id, "DELIVERED")
	} else {
		s.mq.PublishStatus(order.Id, "CANCELED")
	}
	return err
}

func (s *SrvDelivery) RevertDeliveryOperation(idOrder string, statusOrder string) error {
	if statusOrder != "CANCELED" {
		return nil
	}
	operation, err := s.storage.GetDeliveryOperation(idOrder)
	if err != nil {
		return err
	}
	if len(operation.Id) == 0 {
		return ErrDeliveryIdOrderNotFound
	}
	operation.Id = uuid.New().String()
	operation.Operation = "CANCELED"
	err = s.storage.CreateDeliveryOperation(operation)
	if err != nil {
		return err
	}
	return nil
}
