package pay_amqp

import (
	"context"
	"encoding/json"
	"fmt"

	amqp_pub "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/amqp/pub"
	amqp_settings "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/amqp/settings"
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

type SrvPay interface {
	CreatePaymentOperation(order order_app.Order) error
}

type SrvPayAMQP struct {
	logger Logger
	srvPay SrvPay
	pub    amqp_pub.AmqpPub
	uri    string
}

func New(logger Logger, srvPay SrvPay, uri string) *SrvPayAMQP {
	return &SrvPayAMQP{
		logger: logger,
		srvPay: srvPay,
		pub:    *amqp_pub.New(logger),
		uri:    uri,
	}
}

func (a *SrvPayAMQP) Start(ctx context.Context) error {
	conn, err := amqp.Dial(a.uri)
	if err != nil {
		return err
	}
	c := amqp_sub.New("SrvPayAMQP", conn, a.logger)
	msgs, err := c.Consume(ctx, amqp_settings.QueueOrder, amqp_settings.ExchangeOrder,
		"direct", amqp_settings.RoutingKeyPayService)
	if err != nil {
		return err
	}

	a.logger.Info("start consuming...")

	for m := range msgs {
		order := order_app.Order{}
		json.Unmarshal(m.Data, &order)
		if err != nil {
			return err
		}
		a.logger.Info(fmt.Sprintf("receive new message:%+v\n", order))

		err := a.srvPay.CreatePaymentOperation(order)
		if err == nil {
			a.publishStatus(order.Id, "PAYED")
			a.publishOrder(order)
		} else {
			a.publishStatus(order.Id, "CANCELED")
		}
	}
	return nil
}

func (a *SrvPayAMQP) publishOrder(order order_app.Order) error {
	orderStr, err := json.Marshal(order)
	if err != nil {
		return err
	}
	if err := a.pub.Publish(a.uri, amqp_settings.ExchangeOrder, "direct",
		amqp_settings.RoutingKeyStoreService, string(orderStr), true); err != nil {
		return err
	}
	a.logger.Info("publish order for queue is OK ( OrderId: " + order.Id + ")")
	return nil
}

func (a *SrvPayAMQP) publishStatus(idOrder string, statusOrder string) error {
	orderStatusEvent := order_app.OrderEvent{Id: idOrder, Status: statusOrder}
	orderStatusStr, err := json.Marshal(orderStatusEvent)
	if err != nil {
		return err
	}
	if err := a.pub.Publish(a.uri, amqp_settings.ExchangeStatus, "direct",
		"", string(orderStatusStr), true); err != nil {
		return err
	}
	a.logger.Info("publish order status for queue is OK ( OrderId: " + orderStatusEvent.Id + ")")
	return nil
}
