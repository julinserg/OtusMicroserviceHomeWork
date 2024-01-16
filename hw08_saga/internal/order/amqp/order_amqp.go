package order_amqp

import (
	"context"
	"encoding/json"
	"fmt"

	amqp_pub "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/amqp/pub"
	amqp_sub "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/amqp/sub"
	order_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/order/app"
	"github.com/streadway/amqp"
)

const exchangeOrder = "exchange_order"
const exchangeStatus = "exchange_status"
const queueStatus = "queue_status"

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
	Warn(msg string)
}

type Storage interface {
	UpdateOrderStatus(idOrder string, status string) error
}

type SrvOrderAMQP struct {
	logger  Logger
	storage Storage
	pub     amqp_pub.AmqpPub
	uri     string
}

func New(logger Logger, storage Storage, uri string) *SrvOrderAMQP {
	return &SrvOrderAMQP{
		logger:  logger,
		storage: storage,
		pub:     *amqp_pub.New(logger),
		uri:     uri,
	}
}

func (a *SrvOrderAMQP) Start(ctx context.Context) error {
	conn, err := amqp.Dial(a.uri)
	if err != nil {
		return err
	}

	c := amqp_sub.New("SrvOrderAMQP", conn, a.logger)
	msgs, err := c.Consume(ctx, queueStatus, exchangeStatus, "direct", "")
	if err != nil {
		return err
	}

	err = a.pub.CreateExchange(a.uri, exchangeOrder, "direct")
	if err != nil {
		return err
	}

	a.logger.Info("start consuming...")

	for m := range msgs {
		notifyEvent := order_app.OrderEvent{}
		json.Unmarshal(m.Data, &notifyEvent)
		if err != nil {
			return err
		}
		a.logger.Info(fmt.Sprintf("receive new message:%+v\n", notifyEvent))
		a.storage.UpdateOrderStatus(notifyEvent.Id, notifyEvent.Status)
	}
	return nil
}

func (a *SrvOrderAMQP) Publish(order order_app.Order) error {
	orderStr, err := json.Marshal(order)
	if err != nil {
		return err
	}
	if err := a.pub.Publish(a.uri, exchangeOrder, "direct", "forPayService", string(orderStr), true); err != nil {
		return err
	}
	a.logger.Info("publish order for queue is OK ( OrderId: " + order.Id + ")")
	return nil
}
