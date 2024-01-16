package amqp_sub

import (
	"context"
	"fmt"

	"github.com/streadway/amqp"
)

type RMQConnection interface {
	Channel() (*amqp.Channel, error)
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
	Warn(msg string)
}

type AmqpSub struct {
	name   string
	logger Logger
	conn   RMQConnection
}

func New(name string, conn RMQConnection, logger Logger) *AmqpSub {
	return &AmqpSub{
		name:   name,
		conn:   conn,
		logger: logger,
	}
}

type Message struct {
	Ctx  context.Context
	Data []byte
}

func (amqpSub *AmqpSub) Consume(ctx context.Context, queueName string,
	exchangeName string, exchangeType string, key string) (<-chan Message, error) {
	messages := make(chan Message)

	ch, err := amqpSub.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	if err = ch.ExchangeDeclare(
		exchangeName, // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	amqpSub.logger.Info("declared Exchange, declaring Queue " + queueName)
	queue, err := ch.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	amqpSub.logger.Info(fmt.Sprintf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key))

	if err = ch.QueueBind(
		queue.Name,   // name of the queue
		key,          // bindingKey
		exchangeName, // sourceExchange
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	go func() {
		<-ctx.Done()
		if err := ch.Close(); err != nil {
			amqpSub.logger.Error(err.Error())
		}
	}()
	amqpSub.logger.Info(fmt.Sprintf("Queue bound to Exchange, starting Consume (consumer tag %q)", amqpSub.name))
	deliveries, err := ch.Consume(queue.Name, amqpSub.name, false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("start consuming: %w", err)
	}

	go func() {
		defer func() {
			close(messages)
			amqpSub.logger.Info("close messages channel")
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case del := <-deliveries:
				if err := del.Ack(false); err != nil {
					amqpSub.logger.Error(err.Error())
				}

				msg := Message{
					Ctx:  context.TODO(),
					Data: del.Body,
				}

				select {
				case <-ctx.Done():
					return
				case messages <- msg:
				}
			}
		}
	}()

	return messages, nil
}
