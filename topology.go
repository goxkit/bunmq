// Copyright (c) 2023, The GoKit Authors
// MIT License
// All rights reserved.

package crmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type (
	// Topology defines the interface for managing RabbitMQ topology components.
	// Topology includes the configuration of exchanges, queues, and their bindings.
	// It provides methods to declare and apply a complete messaging topology.
	Topology interface {
		// Queue adds a queue definition to the topology.
		Queue(q *QueueDefinition) Topology

		// Queues adds multiple queue definitions to the topology.
		Queues(queues []*QueueDefinition) Topology

		// Exchange adds an exchange definition to the topology.
		Exchange(e *ExchangeDefinition) Topology

		// Exchanges adds multiple exchange definitions to the topology.
		Exchanges(e []*ExchangeDefinition) Topology

		// ExchangeBinding adds an exchange-to-exchange binding to the topology.
		ExchangeBinding(b *ExchangeBindingDefinition) Topology

		// QueueBinding adds an exchange-to-queue binding to the topology.
		QueueBinding(b *QueueBindingDefinition) Topology

		// GetQueuesDefinition returns all queue definitions in the topology.
		GetQueuesDefinition() map[string]*QueueDefinition

		// GetQueueDefinition retrieves a queue definition by name.
		// Returns an error if the queue definition doesn't exist.
		GetQueueDefinition(queueName string) (*QueueDefinition, error)

		// Apply declares all the exchanges, queues, and bindings defined in the topology.
		// Returns an error if any part of the topology cannot be applied.
		Apply() error
	}

	// topology is the concrete implementation of the Topology interface.
	// It maintains collections of exchanges, queues, and their bindings,
	// and provides methods to declare and apply them to a RabbitMQ broker.
	topology struct {
		connectionString string
		conn             RMQConnection
		channel          AMQPChannel
		queues           map[string]*QueueDefinition
		queuesBinding    map[string]*QueueBindingDefinition
		exchanges        []*ExchangeDefinition
		exchangesBinding []*ExchangeBindingDefinition
	}
)

// NewTopology creates a new topology instance with the provided configuration.
// It initializes empty collections for queues and queue bindings.
func NewTopology(connectionString string) *topology {
	return &topology{
		connectionString: connectionString,
		queues:           map[string]*QueueDefinition{},
		queuesBinding:    map[string]*QueueBindingDefinition{}}
}

// Queue adds a queue definition to the topology.
// The queue is indexed by its name for easy retrieval.
func (t *topology) Queue(q *QueueDefinition) *topology {
	t.queues[q.name] = q
	return t
}

// Queues adds multiple queue definitions to the topology.
// Each queue is indexed by its name for easy retrieval.
func (t *topology) Queues(queues []*QueueDefinition) *topology {
	for _, q := range queues {
		t.queues[q.name] = q
	}

	return t
}

// GetQueuesDefinition returns all queue definitions in the topology.
func (t *topology) GetQueuesDefinition() map[string]*QueueDefinition {
	return t.queues
}

// GetQueueDefinition retrieves a queue definition by name.
// Returns an error if the queue definition doesn't exist.
func (t *topology) GetQueueDefinition(queueName string) (*QueueDefinition, error) {
	if d, ok := t.queues[queueName]; ok {
		return d, nil
	}

	return nil, NotFoundQueueDefinitionError
}

// Exchange adds an exchange definition to the topology.
func (t *topology) Exchange(e *ExchangeDefinition) *topology {
	t.exchanges = append(t.exchanges, e)
	return t
}

// Exchanges adds multiple exchange definitions to the topology.
func (t *topology) Exchanges(e []*ExchangeDefinition) *topology {
	t.exchanges = append(t.exchanges, e...)
	return t
}

// ExchangeBinding adds an exchange-to-exchange binding to the topology.
func (t *topology) ExchangeBinding(b *ExchangeBindingDefinition) *topology {
	t.exchangesBinding = append(t.exchangesBinding, b)
	return t
}

// QueueBinding adds an exchange-to-queue binding to the topology.
// The binding is indexed by the queue name.
func (t *topology) QueueBinding(b *QueueBindingDefinition) *topology {
	t.queuesBinding[b.queue] = b
	return t
}

// Apply declares all the exchanges, queues, and bindings defined in the topology.
// It follows a specific order to ensure proper dependency resolution:
//  1. Exchanges are declared first
//  2. Queues are declared next (including any associated retry or DLQ queues)
//  3. Queues are bound to exchanges
//  4. Exchanges are bound to other exchanges
//
// This ordering ensures that all required resources exist before binding them together.
// When declaring queues, any retry queues and dead-letter queues specified in the queue
// definitions are automatically created with appropriate arguments for message routing.
//
// Returns the topology and an error if any part of the topology cannot be applied.
// Common error cases include:
//   - Channel is nil (NullableChannelError)
//   - Exchange declaration failure (permission issues, invalid arguments)
//   - Queue declaration failure (permission issues, invalid arguments)
//   - Binding failure (non-existent queues or exchanges)
func (t *topology) Apply() (*topology, error) {
	conn, ch, err := NewConnection(t.connectionString)
	if err != nil {
		return nil, err
	}

	t.conn = conn
	t.channel = ch

	if err := t.declareExchanges(); err != nil {
		return nil, err
	}

	if err := t.declareQueues(); err != nil {
		return nil, err
	}

	if err := t.bindQueues(); err != nil {
		return nil, err
	}

	return t, t.bindExchanges()
}

// declareExchanges declares all the exchanges defined in the topology.
func (t *topology) declareExchanges() error {
	logrus.Debug(LogMessage("declaring exchanges..."))

	for _, exch := range t.exchanges {
		if err := t.channel.ExchangeDeclare(exch.name, exch.kind.String(), exch.durable, exch.delete, false, false, exch.params); err != nil {
			return err
		}
	}

	logrus.Debug(LogMessage("exchanges declared"))

	return nil
}

// declareQueues declares all the queues defined in the topology.
// For each queue, it also declares any associated retry or dead letter queues
// as defined in the queue properties.
func (t *topology) declareQueues() error {
	logrus.Debug(LogMessage("declaring queues..."))
	for _, queue := range t.queues {
		if queue.withRetry {
			logrus.Debug(LogMessage("declaring retry queue..."))

			//queue.RetryName(), true, false, false, false, amqpDlqDeclarationOpts
			if _, err := t.channel.QueueDeclare(queue.RetryName(), queue.durable, queue.delete, queue.exclusive, false, amqp.Table{
				"x-dead-letter-exchange":    "",
				"x-dead-letter-routing-key": queue.name,
				"x-message-ttl":             queue.retryTTL.Milliseconds(),
			}); err != nil {
				return err
			}

			logrus.Debug(LogMessage("retry queue declared"))
		}

		var amqpDlqDeclarationOpts amqp.Table
		if queue.withDLQ && queue.withRetry {
			amqpDlqDeclarationOpts = amqp.Table{
				"x-dead-letter-exchange":    "",
				"x-dead-letter-routing-key": queue.RetryName(),
			}
		}

		if queue.withDLQ && !queue.withRetry {
			amqpDlqDeclarationOpts = amqp.Table{
				"x-dead-letter-exchange":    "",
				"x-dead-letter-routing-key": queue.DLQName(),
			}
		}

		if queue.withDLQ {
			logrus.Debug(LogMessage("declaring dlq queue..."))

			//queue.DLQName(), true, false, false, false, amqpDlqDeclarationOpts
			if _, err := t.channel.QueueDeclare(queue.DLQName(), queue.durable, queue.delete, queue.exclusive, false, amqpDlqDeclarationOpts); err != nil {
				return err
			}

			logrus.Debug(LogMessage("dlq queue declared"))
		}

		if _, err := t.channel.QueueDeclare(queue.name, queue.durable, queue.delete, queue.exclusive, false, amqpDlqDeclarationOpts); err != nil {
			return err
		}
	}
	logrus.Debug(LogMessage("queues declared"))
	return nil
}

// bindQueues binds all the queues to their respective exchanges
// according to the queue bindings defined in the topology.
func (t *topology) bindQueues() error {
	logrus.Debug(LogMessage("binding queues..."))

	for _, bind := range t.queuesBinding {
		if err := t.channel.QueueBind(bind.queue, bind.routingKey, bind.exchange, false, bind.args); err != nil {
			return err
		}
	}

	logrus.Debug(LogMessage("queue bonded"))

	return nil
}

// bindExchanges binds exchanges to each other according to
// the exchange bindings defined in the topology.
func (t *topology) bindExchanges() error {
	logrus.Debug(LogMessage("binding exchanges..."))

	for _, bind := range t.exchangesBinding {
		if err := t.channel.ExchangeBind(bind.destination, bind.routingKey, bind.source, false, bind.args); err != nil {
			return err
		}
	}

	logrus.Debug(LogMessage("exchanges bonded"))

	return nil
}
