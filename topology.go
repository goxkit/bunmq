// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"sync"

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
		Apply() (ConnectionManager, error)

		// GetConnectionManager returns the underlying connection manager
		GetConnectionManager() ConnectionManager

		// ApplyWithReconnection applies the topology with automatic reconnection support
		ApplyWithReconnection(config ...ReconnectionConfig) (ConnectionManager, error)

		// IsHealthy checks if the underlying connections are healthy
		IsHealthy() bool

		// Reconnect manually triggers a reconnection
		Reconnect() error
	}

	// resilientTopology implements ResilientTopology with connection management
	topology struct {
		connectionString  string
		channel           AMQPChannel
		queues            map[string]*QueueDefinition
		queuesBinding     map[string]*QueueBindingDefinition
		exchanges         []*ExchangeDefinition
		exchangesBinding  []*ExchangeBindingDefinition
		connectionManager ConnectionManager
		mu                sync.RWMutex
		applied           bool
	}
)

// NewResilientTopology creates a new resilient topology with automatic reconnection
func NewTopology(connectionString string) Topology {
	return &topology{
		connectionString: connectionString,
		queues:           map[string]*QueueDefinition{},
		queuesBinding:    map[string]*QueueBindingDefinition{},
		exchanges:        []*ExchangeDefinition{},
		exchangesBinding: []*ExchangeBindingDefinition{},
	}
}

// Queue adds a queue definition to the topology (chainable)
func (t *topology) Queue(q *QueueDefinition) Topology {
	t.queues[q.name] = q
	return t
}

// Queues adds multiple queue definitions to the topology (chainable)
func (t *topology) Queues(queues []*QueueDefinition) Topology {
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

// Exchange adds an exchange definition to the topology (chainable)
func (t *topology) Exchange(e *ExchangeDefinition) Topology {
	t.exchanges = append(t.exchanges, e)
	return t
}

// Exchanges adds multiple exchange definitions to the topology (chainable)
func (t *topology) Exchanges(e []*ExchangeDefinition) Topology {
	t.exchanges = append(t.exchanges, e...)
	return t
}

// ExchangeBinding adds an exchange-to-exchange binding to the topology (chainable)
func (t *topology) ExchangeBinding(b *ExchangeBindingDefinition) Topology {
	t.exchangesBinding = append(t.exchangesBinding, b)
	return t
}

// QueueBinding adds an exchange-to-queue binding to the topology (chainable)
func (t *topology) QueueBinding(b *QueueBindingDefinition) Topology {
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
func (t *topology) Apply() (ConnectionManager, error) {
	manager, err := NewConnectionManager(t.connectionString)
	if err != nil {
		return nil, err
	}

	ch, err := manager.GetChannel()
	if err != nil {
		return nil, err
	}

	t.channel = ch
	t.connectionManager = manager

	if err := t.declareExchanges(); err != nil {
		return nil, err
	}

	if err := t.declareQueues(); err != nil {
		return nil, err
	}

	if err := t.bindQueues(); err != nil {
		return nil, err
	}

	if err := t.bindExchanges(); err != nil {
		return nil, err
	}

	return t.connectionManager, nil
}

// declareExchanges declares all the exchanges defined in the topology.
func (t *topology) declareExchanges() error {
	logrus.Info("bunmq declaring exchanges...")

	for _, exch := range t.exchanges {
		if err := t.channel.ExchangeDeclare(exch.name, exch.kind.String(), exch.durable, exch.delete, false, false, exch.params); err != nil {
			logrus.WithError(err).Errorf("bunmq failure to declare exchange: %s", exch.name)
			return err
		}
	}

	logrus.Info("bunmq exchanges declared")

	return nil
}

// declareQueues declares all the queues defined in the topology.
// For each queue, it also declares any associated retry or dead letter queues
// as defined in the queue properties.
func (t *topology) declareQueues() error {
	logrus.Info("bunmq declaring queues...")

	for _, queue := range t.queues {
		logrus.Infof("bunmq declaring queue: %s ...", queue.name)

		if queue.withRetry {
			logrus.Infof("bunmq declaring retry queue: %s ...", queue.RetryName())

			//queue.RetryName(), true, false, false, false, amqpDlqDeclarationOpts
			if _, err := t.channel.QueueDeclare(queue.RetryName(), queue.durable, queue.delete, queue.exclusive, false, amqp.Table{
				"x-dead-letter-exchange":    "",
				"x-dead-letter-routing-key": queue.name,
				"x-message-ttl":             queue.retryTTL.Milliseconds(),
				"x-retry-count":             queue.retries,
			}); err != nil {
				return err
			}

			logrus.Info("bunmq retry queue declared")
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

		if queue.withDLQMaxLength {
			amqpDlqDeclarationOpts["x-max-length"] = queue.dlqMaxLength
		}

		if queue.withDLQ {
			logrus.Infof("bunmq declaring dlq queue: %s ...", queue.DLQName())

			//queue.DLQName(), true, false, false, false, amqpDlqDeclarationOpts
			if _, err := t.channel.QueueDeclare(queue.DLQName(), queue.durable, queue.delete, queue.exclusive, false, amqpDlqDeclarationOpts); err != nil {
				logrus.WithError(err).Errorf("bunmq failure to declare dlq queue: %s", queue.DLQName())
				return err
			}

			delete(amqpDlqDeclarationOpts, "x-max-length")
			logrus.Info("bunmq dlq queue declared")
		}

		if queue.withMaxLength {
			amqpDlqDeclarationOpts["x-max-length"] = queue.maxLength
		}

		if _, err := t.channel.QueueDeclare(queue.name, queue.durable, queue.delete, queue.exclusive, false, amqpDlqDeclarationOpts); err != nil {
			logrus.WithError(err).Errorf("bunmq failure to declare queue: %s", queue.name)
			return err
		}
	}

	logrus.Info("bunmq queues declared")
	return nil
}

// bindQueues binds all the queues to their respective exchanges
// according to the queue bindings defined in the topology.
func (t *topology) bindQueues() error {
	logrus.Info("bunmq binding queues...")

	for _, bind := range t.queuesBinding {
		if err := t.channel.QueueBind(bind.queue, bind.routingKey, bind.exchange, false, bind.args); err != nil {
			logrus.WithError(err).Errorf("bunmq failure to bind queue: %s to exchange: %s", bind.queue, bind.exchange)
			return err
		}
	}

	logrus.Info("bunmq queues bonded")

	return nil
}

// bindExchanges binds exchanges to each other according to
// the exchange bindings defined in the topology.
func (t *topology) bindExchanges() error {
	logrus.Info("bunmq binding exchanges...")

	for _, bind := range t.exchangesBinding {
		if err := t.channel.ExchangeBind(bind.destination, bind.routingKey, bind.source, false, bind.args); err != nil {
			logrus.WithError(err).Errorf("bunmq failure to bind exchange: %s to %s", bind.destination, bind.source)
			return err
		}
	}

	logrus.Info("bunmq exchanges bonded")

	return nil
}

// ApplyWithReconnection applies the topology with automatic reconnection support
func (t *topology) ApplyWithReconnection(config ...ReconnectionConfig) (ConnectionManager, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Create connection manager
	cm, err := NewConnectionManager(t.connectionString, config...)
	if err != nil {
		return nil, err
	}

	t.connectionManager = cm

	// Set up reconnection callback to reapply topology
	cm.SetReconnectCallback(t.reapplyTopology)

	// Apply initial topology
	if err := t.applyTopologyToManager(); err != nil {
		cm.Close()
		return nil, err
	}

	t.applied = true
	return cm, nil
}

// reapplyTopology is called when reconnection occurs
func (t *topology) reapplyTopology(conn RMQConnection, ch AMQPChannel) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.applied {
		return
	}

	logrus.Info("bunmq reapplying topology after reconnection...")

	// Update topology's connection and channel
	t.channel = ch

	// Reapply topology declarations
	if err := t.declareExchanges(); err != nil {
		logrus.WithError(err).Error("bunmq failed to redeclare exchanges")
		return
	}

	if err := t.declareQueues(); err != nil {
		logrus.WithError(err).Error("bunmq failed to redeclare queues")
		return
	}

	if err := t.bindQueues(); err != nil {
		logrus.WithError(err).Error("bunmq failed to rebind queues")
		return
	}

	if err := t.bindExchanges(); err != nil {
		logrus.WithError(err).Error("bunmq failed to rebind exchanges")
		return
	}

	logrus.Info("bunmq topology reapplied successfully")
}

// applyTopologyToManager applies the topology using the connection manager
func (t *topology) applyTopologyToManager() error {
	ch, err := t.connectionManager.GetChannel()
	if err != nil {
		return err
	}

	// Update topology's connection and channel
	t.channel = ch

	// Apply topology declarations
	if err := t.declareExchanges(); err != nil {
		return err
	}

	if err := t.declareQueues(); err != nil {
		return err
	}

	if err := t.bindQueues(); err != nil {
		return err
	}

	if err := t.bindExchanges(); err != nil {
		return err
	}

	return nil
}

// GetConnectionManager returns the underlying connection manager
func (t *topology) GetConnectionManager() ConnectionManager {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connectionManager
}

// IsHealthy checks if the underlying connections are healthy
func (t *topology) IsHealthy() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.connectionManager == nil {
		return false
	}

	return t.connectionManager.IsHealthy()
}

// Reconnect manually triggers a reconnection
func (t *topology) Reconnect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connectionManager == nil {
		return NewBunMQError("connection manager not initialized")
	}

	// Close current connection to trigger reconnection
	if err := t.connectionManager.Close(); err != nil {
		logrus.WithError(err).Error("bunmq error closing connection manager for reconnection")
	}

	// Create new connection manager
	cm, err := NewConnectionManager(t.connectionManager.GetConnectionString())
	if err != nil {
		return err
	}

	t.connectionManager = cm
	cm.SetReconnectCallback(t.reapplyTopology)

	// Reapply topology
	return t.applyTopologyToManager()
}
