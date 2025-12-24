// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type (
	// AMQPChannel defines the interface for a RabbitMQ channel.
	// It abstracts the operations that can be performed on a channel such as
	// declaring exchanges and queues, binding them, and publishing or consuming messages.
	AMQPChannel interface {
		// ExchangeDeclare declares an exchange on the channel.
		// The exchange will be created if it doesn't already exist.
		// Parameters:
		//   - name: The name of the exchange
		//   - kind: The exchange type (direct, fanout, topic, headers)
		//   - durable: Survive broker restarts
		//   - autoDelete: Delete when no longer used
		//   - internal: Can only be published to by other exchanges
		//   - noWait: Don't wait for a server confirmation
		//   - args: Additional arguments
		ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error

		// ExchangeBind binds an exchange to another exchange.
		// Parameters:
		//   - destination: The name of the destination exchange
		//   - key: The routing key to use
		//   - source: The name of the source exchange
		//   - noWait: Don't wait for a server confirmation
		//   - args: Additional arguments
		ExchangeBind(destination, key, source string, noWait bool, args amqp.Table) error

		// QueueDeclare declares a queue on the channel.
		// The queue will be created if it doesn't already exist.
		// Parameters:
		//   - name: The name of the queue
		//   - durable: Survive broker restarts
		//   - autoDelete: Delete when no longer used
		//   - exclusive: Used by only one connection and deleted when that connection closes
		//   - noWait: Don't wait for a server confirmation
		//   - args: Additional arguments
		// Returns the queue and any error encountered.
		QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)

		// QueueBind binds a queue to an exchange.
		// Parameters:
		//   - name: The name of the queue
		//   - key: The routing key to use
		//   - exchange: The name of the exchange
		//   - noWait: Don't wait for a server confirmation
		//   - args: Additional arguments
		QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error

		// Consume starts delivering messages from a queue.
		// Parameters:
		//   - queue: The name of the queue
		//   - consumer: The consumer tag (empty string to have the server generate one)
		//   - autoAck: Acknowledge messages automatically when delivered
		//   - exclusive: Request exclusive consumer access
		//   - noLocal: Don't deliver messages published on this connection
		//   - noWait: Don't wait for a server confirmation
		//   - args: Additional arguments
		// Returns a channel of delivered messages and any error encountered.
		Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)

		// Publish publishes a message to an exchange.
		// Parameters:
		//   - exchange: The name of the exchange
		//   - key: The routing key to use
		//   - mandatory: Return message if it can't be routed to a queue
		//   - immediate: Return message if it can't be delivered to a consumer immediately
		//   - msg: The message to publish
		Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error

		PublishWithDeferredConfirm(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) (*amqp.DeferredConfirmation, error)

		// IsClosed checks if the channel is closed.
		IsClosed() bool

		// Close closes the channel gracefully.
		Close() error

		// NotifyClose returns a channel that receives notifications when the channel is closed.
		// This is essential for connection management and automatic reconnection strategies.
		NotifyClose(receiver chan *amqp.Error) chan *amqp.Error

		// NotifyCancel returns a channel that receives notifications when a consumer is cancelled.
		// This helps detect when the server cancels consumers due to various conditions.
		NotifyCancel(receiver chan string) chan string

		/*
			Qos controls how many messages or how many bytes the server will try to keep on
			the network for consumers before receiving delivery acks.  The intent of Qos is
			to make sure the network buffers stay full between the server and client.

			With a prefetch count greater than zero, the server will deliver that many
			messages to consumers before acknowledgments are received.  The server ignores
			this option when consumers are started with noAck because no acknowledgments
			are expected or sent.

			With a prefetch size greater than zero, the server will try to keep at least
			that many bytes of deliveries flushed to the network before receiving
			acknowledgments from the consumers.  This option is ignored when consumers are
			started with noAck.

			When global is true, these Qos settings apply to all existing and future
			consumers on all channels on the same connection.  When false, the Channel.Qos
			settings will apply to all existing and future consumers on this channel.

			Please see the RabbitMQ Consumer Prefetch documentation for an explanation of
			how the global flag is implemented in RabbitMQ, as it differs from the
			AMQP 0.9.1 specification in that global Qos settings are limited in scope to
			channels, not connections (https://www.rabbitmq.com/consumer-prefetch.html).

			To get round-robin behavior between consumers consuming from the same queue on
			different connections, set the prefetch count to 1, and the next available
			message on the server will be delivered to the next available consumer.

			If your consumer work time is reasonably consistent and not much greater
			than two times your network round trip time, you will see significant
			throughput improvements starting with a prefetch count of 2 or slightly
			greater as described by benchmarks on RabbitMQ.

			http://www.rabbitmq.com/blog/2012/04/25/rabbitmq-performance-measurements-part-2/
		*/
		Qos(prefetchCount, prefetchSize int, global bool) error
	}
)

// dial is a variable that holds the function to establish a connection to RabbitMQ.
// It allows for mocking in tests.
var dial = func(appName, connectionString string) (RMQConnection, error) {
	return amqp.DialConfig(connectionString, amqp.Config{
		Properties: amqp.Table{
			"connection_name": fmt.Sprintf("%d-%s", time.Now().Unix(), appName),
		},
	})
}

// NewConnection creates a new RabbitMQ connection and channel.
// It establishes a connection to the RabbitMQ server using the provided configuration,
// then creates a channel on that connection.
// Returns the connection, channel, and any error encountered.
func NewConnection(appName, connectionString string) (RMQConnection, AMQPChannel, error) {
	logrus.Debug("bunmq connecting to rabbitmq...")
	conn, err := dial(appName, connectionString)
	if err != nil {
		logrus.WithError(err).Error("bunmq failure to connect to the broker")
		return nil, nil, rabbitMQDialError(err)
	}
	logrus.Debug("bunmq connected to rabbitmq")

	logrus.Debug("bunmq creating amqp channel...")
	ch, err := conn.Channel()
	if err != nil {
		logrus.WithError(err).Error("bunmq failure to establish the channel")
		return nil, nil, getChannelError(err)
	}
	logrus.Debug("bunmq created amqp channel")

	return conn, ch, nil
}
