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
	logger           Logger
	storage          Storage
	pub              amqp_pub.AmqpPub
	uri              string
	consumer         string
	queue            string
	exchange         string
	exchangeType     string
	routingKey       string
	exchangeUser     string
	exchangeUserType string
}

func New(logger Logger, storage Storage, uri string, consumer string,
	queue string, exchange string, exchangeType string,
	routingKey string, exchangeUser string, exchangeUserType string) *SrvOrderAMQP {
	return &SrvOrderAMQP{
		logger:           logger,
		storage:          storage,
		pub:              *amqp_pub.New(logger),
		uri:              uri,
		consumer:         consumer,
		queue:            queue,
		exchange:         exchange,
		exchangeType:     exchangeType,
		routingKey:       routingKey,
		exchangeUser:     exchangeUser,
		exchangeUserType: exchangeUserType,
	}
}

func (a *SrvOrderAMQP) Start(ctx context.Context) error {
	conn, err := amqp.Dial(a.uri)
	if err != nil {
		return err
	}

	c := amqp_sub.New(a.consumer, conn, a.logger)
	msgs, err := c.Consume(ctx, a.queue, a.exchange, a.exchangeType, a.routingKey)
	if err != nil {
		return err
	}

	err = a.pub.CreateExchange(a.uri, a.exchangeUser, a.exchangeUserType)
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
	if err := a.pub.Publish(a.uri, a.exchangeUser, a.exchangeUserType, "", string(orderStr), true); err != nil {
		return err
	}
	a.logger.Info("publish order for queue is OK ( OrderId: " + order.Id + ")")
	return nil
}
