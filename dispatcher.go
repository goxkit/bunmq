// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	ConsumerDefinitionByType ConsumerDefinitionType = iota + 1
	ConsumerDefinitionByExchange
	ConsumerDefinitionByRoutingKey
	ConsumerDefinitionByExchangeRoutingKey
)

type (
	// Dispatcher defines an interface for managing RabbitMQ message consumption.
	// It provides methods to register message handlers and consume messages in a blocking manner.
	Dispatcher interface {
		// RegisterByType associates a queue with a message type and a handler function.
		// It ensures that messages from the specified queue are processed by the handler.
		// Returns an error if the registration parameters are invalid or if the queue definition is not found.
		RegisterByType(queue string, typE any, handler ConsumerHandler) error

		// RegisterByExchange associates a queue with a message handler based on the exchange.
		// It ensures that all messages from the specified exchange are processed by the handler.
		// Returns an error if the registration parameters are invalid or if the queue definition is not found.
		RegisterByExchange(queue string, msg any, exchange string, handler ConsumerHandler) error

		// RegisterByRoutingKey associates a queue with a message handler based on the routing key.
		// It ensures that all messages with the specified routing key are processed by the handler.
		// Returns an error if the registration parameters are invalid or if the queue definition is not found.
		RegisterByRoutingKey(queue string, msg any, routingKey string, handler ConsumerHandler) error

		// RegisterByExchangeRoutingKey associates a queue with a message handler based on both exchange and routing key.
		// It ensures that messages matching both the exchange and routing key are processed by the handler.
		// Returns an error if the registration parameters are invalid or if the queue definition is not found.
		RegisterByExchangeRoutingKey(queue string, msg any, exchange, routingKey string, handler ConsumerHandler) error

		// ConsumeBlocking starts consuming messages and dispatches them to the registered handlers.
		// This method blocks execution until the process is terminated by a signal.
		ConsumeBlocking()
	}

	ConsumerDefinitionType int

	// ConsumerHandler is a function type that defines message handler callbacks.
	// It receives a context (for tracing), the unmarshaled message, and metadata about the delivery.
	// Returns an error if the message processing fails.
	ConsumerHandler = func(ctx context.Context, msg any, metadata *DeliveryMetadata) error

	// ConsumerDefinition represents the configuration for a consumer.
	// It holds information about the queue, message type, and handler function.
	ConsumerDefinition struct {
		typ             ConsumerDefinitionType
		queue           string
		exchange        string
		routingKey      string
		msgType         string
		reflect         *reflect.Value
		queueDefinition *QueueDefinition
		handler         ConsumerHandler
	}

	// deliveryMetadata contains metadata extracted from an AMQP delivery.
	// This includes message ID, retry count, message type, and headers.
	DeliveryMetadata struct {
		MessageID      string
		XCount         int64
		Type           string
		OriginExchange string
		RoutingKey     string
		Headers        map[string]interface{}
	}

	// dispatcher is the concrete implementation of the Dispatcher interface.
	// It manages the registration and execution of message handlers for RabbitMQ queues.
	dispatcher struct {
		manager             ConnectionManager
		queueDefinitions    map[string]*QueueDefinition
		consumersDefinition map[string]*ConsumerDefinition
		tracer              trace.Tracer
		signalCh            chan os.Signal
		mu                  sync.RWMutex
		consuming           bool
		consumeChannels     map[string]<-chan amqp.Delivery
	}
)

// NewDispatcher creates a new dispatcher instance with the provided configuration.
// It initializes signal handling and sets up the necessary components for message consumption.
func NewDispatcher(manager ConnectionManager, queueDefinitions []*QueueDefinition) Dispatcher {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	customQueueDefs := make(map[string]*QueueDefinition)
	for _, q := range queueDefinitions {
		customQueueDefs[q.name] = q
	}

	return &dispatcher{
		manager:             manager,
		queueDefinitions:    customQueueDefs,
		consumersDefinition: map[string]*ConsumerDefinition{},
		tracer:              otel.Tracer("bunmq-dispatcher"),
		signalCh:            signalCh,
		consumeChannels:     make(map[string]<-chan amqp.Delivery),
	}
}

// RegisterByType associates a queue with a message type and a handler function.
// It validates the parameters and ensures that the queue definition exists.
// Returns an error if the registration parameters are invalid or if the queue definition is not found.
func (d *dispatcher) RegisterByType(queue string, msg any, handler ConsumerHandler) error {
	if msg == nil || queue == "" {
		logrus.Error("bunmq invalid parameters to register consumer")
		return InvalidDispatchParamsError
	}

	ref := reflect.New(reflect.TypeOf(msg))
	msgType := fmt.Sprintf("%T", msg)

	_, ok := d.consumersDefinition[msgType]
	if ok {
		logrus.Error("bunmq consumer already registered for this message")
		return ConsumerAlreadyRegisteredForTheMessageError
	}

	def, ok := d.queueDefinitions[queue]
	if !ok {
		logrus.Error("bunmq queue definition not found for the given queue")
		return QueueDefinitionNotFoundError
	}

	d.consumersDefinition[msgType] = &ConsumerDefinition{
		typ:             ConsumerDefinitionByType,
		queue:           queue,
		msgType:         msgType,
		reflect:         &ref,
		queueDefinition: def,
		handler:         handler,
	}

	return nil
}

// RegisterByExchange associates a queue with a message type and handler function based on the exchange.
func (d *dispatcher) RegisterByExchange(queue string, msg any, exchange string, handler ConsumerHandler) error {
	if msg == nil || queue == "" || exchange == "" {
		logrus.Error("bunmq invalid parameters to register consumer")
		return InvalidDispatchParamsError
	}

	ref := reflect.New(reflect.TypeOf(msg))
	msgType := fmt.Sprintf("%T", msg)

	_, ok := d.consumersDefinition[msgType]
	if ok {
		logrus.Error("bunmq consumer already registered for this message")
		return ConsumerAlreadyRegisteredForTheMessageError
	}

	def, ok := d.queueDefinitions[queue]
	if !ok {
		logrus.Error("bunmq queue definition not found for the given queue")
		return QueueDefinitionNotFoundError
	}

	d.consumersDefinition[msgType] = &ConsumerDefinition{
		typ:             ConsumerDefinitionByExchange,
		queue:           queue,
		exchange:        exchange,
		msgType:         msgType,
		reflect:         &ref,
		queueDefinition: def,
		handler:         handler,
	}

	return nil
}

// RegisterByRoutingKey associates a queue with a message type and handler function based on the routing key.
func (d *dispatcher) RegisterByRoutingKey(queue string, msg any, routingKey string, handler ConsumerHandler) error {
	if msg == nil || queue == "" || routingKey == "" {
		logrus.Error("bunmq invalid parameters to register consumer")
		return InvalidDispatchParamsError
	}

	ref := reflect.New(reflect.TypeOf(msg))
	msgType := fmt.Sprintf("%T", msg)

	_, ok := d.consumersDefinition[msgType]
	if ok {
		logrus.Error("bunmq consumer already registered for this message")
		return ConsumerAlreadyRegisteredForTheMessageError
	}

	def, ok := d.queueDefinitions[queue]
	if !ok {
		logrus.Error("bunmq queue definition not found for the given queue")
		return QueueDefinitionNotFoundError
	}

	d.consumersDefinition[msgType] = &ConsumerDefinition{
		typ:             ConsumerDefinitionByRoutingKey,
		queue:           queue,
		routingKey:      routingKey,
		msgType:         msgType,
		reflect:         &ref,
		queueDefinition: def,
		handler:         handler,
	}

	return nil
}

// RegisterByExchangeRoutingKey associates a queue with a message type and handler function based on both exchange and routing key.
func (d *dispatcher) RegisterByExchangeRoutingKey(queue string, msg any, exchange, routingKey string, handler ConsumerHandler) error {
	if msg == nil || queue == "" || exchange == "" || routingKey == "" {
		logrus.Error("bunmq invalid parameters to register consumer")
		return InvalidDispatchParamsError
	}

	ref := reflect.New(reflect.TypeOf(msg))
	msgType := fmt.Sprintf("%T", msg)

	_, ok := d.consumersDefinition[msgType]
	if ok {
		logrus.Error("bunmq consumer already registered for this message")
		return ConsumerAlreadyRegisteredForTheMessageError
	}

	def, ok := d.queueDefinitions[queue]
	if !ok {
		logrus.Error("bunmq queue definition not found for the given queue")
		return QueueDefinitionNotFoundError
	}

	d.consumersDefinition[msgType] = &ConsumerDefinition{
		typ:             ConsumerDefinitionByExchangeRoutingKey,
		queue:           queue,
		exchange:        exchange,
		routingKey:      routingKey,
		msgType:         msgType,
		reflect:         &ref,
		queueDefinition: def,
		handler:         handler,
	}

	return nil
}

// ConsumeBlocking starts consuming messages from all registered queues.
// It creates a goroutine for each consumer and blocks until a termination signal is received.
// This method should be called after all Register operations are complete.
func (d *dispatcher) ConsumeBlocking() {
	d.mu.Lock()
	d.consuming = true
	d.mu.Unlock()

	logrus.Info("bunmq resilient dispatcher started, waiting for messages...")

	// Start all consumers
	for _, cd := range d.consumersDefinition {
		go d.consume(cd.typ, cd.queue, cd.msgType, cd.exchange, cd.routingKey)
	}

	// Wait for shutdown signal
	<-d.signalCh
	logrus.Info("bunmq signal received, closing resilient dispatcher")

	d.mu.Lock()
	d.consuming = false
	d.mu.Unlock()
}

// consume starts consuming messages from a specific queue with automatic reconnection
func (d *dispatcher) consume(typ ConsumerDefinitionType, queue, msgType, exchange, routingKey string) {
	for {
		// Check if we should stop consuming
		d.mu.RLock()
		if !d.consuming {
			d.mu.RUnlock()
			return
		}
		d.mu.RUnlock()

		// Get current channel from connection manager
		ch, err := d.manager.GetChannel()
		if err != nil {
			logrus.WithError(err).Errorf("bunmq failed to get channel for queue: %s, retrying...", queue)
			time.Sleep(time.Second * 5)
			continue
		}

		// Start consuming from the queue
		delivery, err := ch.Consume(queue, "", false, false, false, false, nil)
		if err != nil {
			logrus.WithError(err).Errorf("bunmq failure to declare consumer for queue: %s, retrying...", queue)
			time.Sleep(time.Second * 5)
			continue
		}

		d.mu.Lock()
		d.consumeChannels[queue] = delivery
		d.mu.Unlock()

		logrus.Infof("bunmq started consuming from queue: %s", queue)

		// Process messages from this consumer
		for received := range delivery {
			// Process the message using the same logic as the original dispatcher
			d.processReceivedMessage(typ, &received)
		}

		// If we reach here, the delivery channel was closed
		logrus.Warnf("bunmq delivery channel closed for queue: %s, will attempt to reconnect", queue)

		d.mu.Lock()
		delete(d.consumeChannels, queue)
		d.mu.Unlock()

		// Wait a bit before trying to reconnect
		logrus.Warnf("bunmq waiting before reconnecting for queue: %s", queue)
		time.Sleep(time.Second * 2)
	}
}

// consume starts consuming messages from a specific queue.
// This internal method is responsible for the core message processing workflow.
func (d *dispatcher) processReceivedMessage(typ ConsumerDefinitionType, received *amqp.Delivery) {
	metadata, err := d.extractMetadata(received)
	if err != nil {
		_ = received.Ack(false)
		return
	}

	def, err := d.extractDefByType(typ, metadata)
	if err != nil {
		logrus.
			WithField("messageID", metadata.MessageID).
			Warnf("bunmq no consumer found for message type: %s", metadata.Type)
		_ = received.Ack(false)
		return
	}

	logrus.
		WithField("messageID", metadata.MessageID).
		Debugf("bunmq received message: %s", metadata.Type)

	ctx, span := NewConsumerSpan(d.tracer, received.Headers, received.Type)

	ptr := def.reflect.Interface()
	if err = json.Unmarshal(received.Body, ptr); err != nil {
		span.RecordError(err)
		logrus.
			WithContext(ctx).
			WithError(err).
			WithField("messageID", metadata.MessageID).
			Errorf("bunmq unmarshal error: %s", received.Type)
		_ = received.Nack(true, false)
		span.End()
		return
	}

	if def.queueDefinition.withRetry && metadata.XCount >= def.queueDefinition.retries {
		logrus.
			WithContext(ctx).
			WithField("messageID", metadata.MessageID).
			Warnf("bunmq message reprocessed to many times, sending to dead letter")
		_ = received.Ack(false)

		if err = d.publishToDlq(ctx, def, received); err != nil {
			span.RecordError(err)
			logrus.
				WithContext(ctx).
				WithError(err).
				WithField("messageID", metadata.MessageID).
				Error("bunmq failure to publish to dlq")
		}

		span.End()
		return
	}

	bunErr := def.handler(ctx, ptr, metadata)
	if bunErr != nil {
		logrus.
			WithContext(ctx).
			WithError(bunErr).
			WithField("messageID", metadata.MessageID).
			Error("bunmq error to process message")

		if def.queueDefinition.withRetry && errors.Is(bunErr, RetryableError) {
			logrus.
				WithContext(ctx).
				WithField("messageID", metadata.MessageID).
				Warn("bunmq send message to process latter")

			_ = received.Nack(false, false)
			span.End()
			return
		}

		if def.queueDefinition.withDLQ {
			span.RecordError(bunErr)
			_ = received.Ack(false)

			if err = d.publishToDlq(ctx, def, received); err != nil {
				span.RecordError(err)
				logrus.
					WithContext(ctx).
					WithError(err).
					WithField("messageID", metadata.MessageID).
					Error("bunmq failure to publish to dlq")
			}

			span.End()
			return
		}

		logrus.
			WithContext(ctx).
			WithError(bunErr).
			WithField("messageID", metadata.MessageID).
			Error("bunmq failure to process message, in queue without DLQ or retry, removing from queue")
		_ = received.Ack(false)
	}

	logrus.
		WithContext(ctx).
		WithField("messageID", metadata.MessageID).
		Debug("bunmq message processed properly")
	_ = received.Ack(true)
	span.SetStatus(codes.Ok, "success")
	span.End()

}

// extractMetadata extracts relevant metadata from an AMQP delivery.
// This includes the message ID, type, and retry count.
// Returns an error if the message has unformatted headers.
func (d *dispatcher) extractMetadata(delivery *amqp.Delivery) (*DeliveryMetadata, error) {
	typ := delivery.Type
	if typ == "" {
		logrus.
			WithField("messageID", delivery.MessageId).
			Error("bunmq unformatted amqp delivery - missing type parameter")
		return nil, ReceivedMessageWithUnformattedHeaderError
	}

	var xCount int64
	if xDeath, ok := delivery.Headers["x-death"]; ok {
		v, _ := xDeath.([]interface{})
		table, _ := v[0].(amqp.Table)
		count, _ := table["count"].(int64)
		xCount = count
	}

	return &DeliveryMetadata{
		MessageID:      delivery.MessageId,
		Type:           typ,
		XCount:         xCount,
		OriginExchange: delivery.Exchange,
		RoutingKey:     delivery.RoutingKey,
		Headers:        delivery.Headers,
	}, nil
}

func (d *dispatcher) extractDefByType(typ ConsumerDefinitionType, metadata *DeliveryMetadata) (*ConsumerDefinition, error) {
	switch typ {
	case ConsumerDefinitionByType:
		def, ok := d.consumersDefinition[metadata.Type]
		if !ok {
			return nil, fmt.Errorf("bunmq no consumer found for message type: %s", metadata.Type)
		}
		return def, nil
	case ConsumerDefinitionByExchange:
		for _, def := range d.consumersDefinition {
			if def.exchange == metadata.OriginExchange {
				return def, nil
			}
		}
	case ConsumerDefinitionByRoutingKey:
		for _, def := range d.consumersDefinition {
			if def.routingKey == metadata.RoutingKey {
				return def, nil
			}
		}
	case ConsumerDefinitionByExchangeRoutingKey:
		for _, def := range d.consumersDefinition {
			if def.exchange == metadata.OriginExchange && def.routingKey == metadata.RoutingKey {
				return def, nil
			}
		}
	default:
		return nil, fmt.Errorf("bunmq unknown consumer definition type: %d", typ)
	}

	return nil, fmt.Errorf("bunmq no consumer definition found for type: %d", typ)
}

// publishToDlq publishes a message to the dead-letter queue.
// It preserves the original message properties and headers.
func (d *dispatcher) publishToDlq(ctx context.Context, definition *ConsumerDefinition, received *amqp.Delivery) error {
	ch, err := d.manager.GetChannel()
	if err != nil {
		logrus.
			WithContext(ctx).
			WithError(err).
			WithField("messageID", received.MessageId).
			Error("bunmq failure to get channel for DLQ")
		return err
	}

	return ch.Publish("", definition.queueDefinition.dqlName, false, false, amqp.Publishing{
		Headers:     received.Headers,
		Type:        received.Type,
		ContentType: received.ContentType,
		MessageId:   received.MessageId,
		UserId:      received.UserId,
		AppId:       received.AppId,
		Body:        received.Body,
	})
}
