// Copyright (c) 2023, The GoKit Authors
// MIT License
// All rights reserved.

package crmq

type (
	// ExchangeBindingDefinition represents a binding between two exchanges.
	// It defines how messages are routed from a source exchange to a destination exchange
	// based on a routing key and optional arguments.
	ExchangeBindingDefinition struct {
		source      string
		destination string
		routingKey  string
		args        map[string]interface{}
	}

	// QueueBindingDefinition represents a binding between an exchange and a queue.
	// It defines how messages are routed from an exchange to a queue
	// based on a routing key and optional arguments.
	QueueBindingDefinition struct {
		routingKey string
		queue      string
		exchange   string
		args       map[string]interface{}
	}
)

// NewExchangeBiding creates a new exchange binding definition.
// This defines how messages are routed between exchanges.
//
// Exchange bindings allow for creating complex routing topologies where messages
// published to one exchange can be routed to another exchange based on routing rules.
// This is useful for scenarios like:
// - Broadcasting messages across multiple systems
// - Implementing fanout patterns from a direct exchange
// - Creating hierarchical routing structures
//
// Example:
//
//	// Define a binding from a direct to a fanout exchange
//	binding := rabbitmq.NewExchangeBiding()
//	binding.Source("orders").Destination("notifications").RoutingKey("new-order")
//
// Note: This binding needs to be added to the Topology with the ExchangeBinding method.
func NewExchangeBiding() *ExchangeBindingDefinition {
	return &ExchangeBindingDefinition{}
}

// NewQueueBinding creates a new queue binding definition.
// This defines how messages are routed from an exchange to a queue.
func NewQueueBinding() *QueueBindingDefinition {
	return &QueueBindingDefinition{}
}

// RoutingKey sets the routing key for this queue binding.
// The routing key is used to filter messages from the exchange to the queue.
func (b *QueueBindingDefinition) RoutingKey(key string) *QueueBindingDefinition {
	b.routingKey = key
	return b
}

// Queue sets the queue name for this binding.
// This is the destination queue that will receive messages from the exchange.
func (b *QueueBindingDefinition) Queue(name string) *QueueBindingDefinition {
	b.queue = name
	return b
}

// Exchange sets the exchange name for this binding.
// This is the source exchange from which messages will be routed to the queue.
func (b *QueueBindingDefinition) Exchange(name string) *QueueBindingDefinition {
	b.exchange = name
	return b
}

// Args sets additional arguments for this queue binding.
// These are passed as arguments when declaring the binding.
func (b *QueueBindingDefinition) Args(args map[string]interface{}) *QueueBindingDefinition {
	b.args = args
	return b
}

// Source sets the source exchange name for this binding.
// This is the exchange from which messages will be routed.
func (b *ExchangeBindingDefinition) Source(name string) *ExchangeBindingDefinition {
	b.source = name
	return b
}

// Destination sets the destination exchange name for this binding.
// This is the exchange to which messages will be routed.
func (b *ExchangeBindingDefinition) Destination(name string) *ExchangeBindingDefinition {
	b.destination = name
	return b
}

// RoutingKey sets the routing key for this exchange binding.
// The routing key is used to filter messages from the source to the destination exchange.
func (b *ExchangeBindingDefinition) RoutingKey(key string) *ExchangeBindingDefinition {
	b.routingKey = key
	return b
}

// Args sets additional arguments for this exchange binding.
func (b *ExchangeBindingDefinition) Args(args map[string]interface{}) *ExchangeBindingDefinition {
	b.args = args
	return b
}
