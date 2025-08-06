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
	ConsumerDefinitionType int

	// ResilientDispatcher extends the basic dispatcher with automatic reconnection capabilities
	Dispatcher interface {
		// RegisterByType associates a queue with a message type and a handler function.
		// It ensures that messages from the specified queue are processed by the handler.
		// Returns an error if the registration parameters are invalid or if the queue definition is not found.
		RegisterByType(queue string, typE any, handler ConsumerHandler) error

		// IsHealthy checks if the underlying connections are healthy
		IsHealthy() bool

		// GetConnectionManager returns the underlying connection manager
		GetConnectionManager() ConnectionManager

		// SetReconnectCallback sets a callback for when reconnection occurs
		SetReconnectCallback(callback func())

		// ConsumeBlocking starts consuming messages and dispatches them to the registered handlers.
		// This method blocks execution until the process is terminated by a signal.
		ConsumeBlocking()
	}

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

	// resilientDispatcher implements ResilientDispatcher with connection management
	dispatcher struct {
		connectionManager   ConnectionManager
		queueDefinitions    map[string]*QueueDefinition
		consumersDefinition map[string]*ConsumerDefinition
		tracer              trace.Tracer
		signalCh            chan os.Signal
		mu                  sync.RWMutex
		reconnectCallback   func()
		consuming           bool
		consumeChannels     map[string]<-chan amqp.Delivery // Track active consumers
	}
)

// NewResilientDispatcher creates a new resilient dispatcher with automatic reconnection
func NewDispatcher(connectionManager ConnectionManager, queueDefinitions []*QueueDefinition) Dispatcher {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	customQueueDefs := make(map[string]*QueueDefinition)
	for _, q := range queueDefinitions {
		customQueueDefs[q.name] = q
	}

	rd := &dispatcher{
		connectionManager:   connectionManager,
		queueDefinitions:    customQueueDefs,
		consumersDefinition: map[string]*ConsumerDefinition{},
		tracer:              otel.Tracer("bunmq-dispatcher"),
		signalCh:            signalCh,
		consumeChannels:     make(map[string]<-chan amqp.Delivery),
	}

	// Set up reconnection callback to restart consumers
	connectionManager.SetReconnectCallback(rd.handleReconnection)

	return rd
}

// handleReconnection is called when the connection manager reconnects
func (rd *dispatcher) handleReconnection(conn RMQConnection, ch AMQPChannel) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	logrus.Info("bunmq dispatcher handling reconnection...")

	// Clear old consume channels as they're now invalid
	rd.consumeChannels = make(map[string]<-chan amqp.Delivery)

	// If we were consuming, restart all consumers
	if rd.consuming {
		go rd.restartConsumers()
	}

	// Call user-defined reconnection callback if set
	if rd.reconnectCallback != nil {
		go rd.reconnectCallback()
	}

	logrus.Info("bunmq dispatcher reconnection handled")
}

// restartConsumers restarts all registered consumers after reconnection
func (rd *dispatcher) restartConsumers() {
	logrus.Info("bunmq restarting consumers after reconnection...")

	for _, cd := range rd.consumersDefinition {
		go rd.consume(cd.typ, cd.queue, cd.msgType, cd.exchange, cd.routingKey)
	}

	logrus.Info("bunmq consumers restarted successfully")
}

// RegisterByType associates a queue with a message type and a handler function
func (rd *dispatcher) RegisterByType(queue string, typE any, handler ConsumerHandler) error {
	return rd.registerConsumer(ConsumerDefinitionByType, queue, typE, "", "", handler)
}

// RegisterByExchange associates a queue with a message handler based on the exchange
func (rd *dispatcher) RegisterByExchange(queue string, msg any, exchange string, handler ConsumerHandler) error {
	return rd.registerConsumer(ConsumerDefinitionByExchange, queue, msg, exchange, "", handler)
}

// RegisterByRoutingKey associates a queue with a message handler based on the routing key
func (rd *dispatcher) RegisterByRoutingKey(queue string, msg any, routingKey string, handler ConsumerHandler) error {
	return rd.registerConsumer(ConsumerDefinitionByRoutingKey, queue, msg, "", routingKey, handler)
}

// RegisterByExchangeRoutingKey associates a queue with a message handler based on both exchange and routing key
func (rd *dispatcher) RegisterByExchangeRoutingKey(queue string, msg any, exchange, routingKey string, handler ConsumerHandler) error {
	return rd.registerConsumer(ConsumerDefinitionByExchangeRoutingKey, queue, msg, exchange, routingKey, handler)
}

// registerConsumer is a helper method to register consumers with proper validation
func (rd *dispatcher) registerConsumer(typ ConsumerDefinitionType, queue string, msg any, exchange, routingKey string, handler ConsumerHandler) error {
	if msg == nil || queue == "" {
		logrus.Error("bunmq invalid parameters to register consumer")
		return InvalidDispatchParamsError
	}

	// Validate type-specific parameters
	switch typ {
	case ConsumerDefinitionByExchange:
		if exchange == "" {
			logrus.Error("bunmq exchange cannot be empty for this registration type")
			return InvalidDispatchParamsError
		}
	case ConsumerDefinitionByRoutingKey:
		if routingKey == "" {
			logrus.Error("bunmq routing key cannot be empty for this registration type")
			return InvalidDispatchParamsError
		}
	case ConsumerDefinitionByExchangeRoutingKey:
		if exchange == "" || routingKey == "" {
			logrus.Error("bunmq exchange and routing key cannot be empty for this registration type")
			return InvalidDispatchParamsError
		}
	}

	ref := reflect.New(reflect.TypeOf(msg))
	msgType := fmt.Sprintf("%T", msg)

	rd.mu.Lock()
	defer rd.mu.Unlock()

	_, ok := rd.consumersDefinition[msgType]
	if ok {
		logrus.Error("bunmq consumer already registered for this message")
		return ConsumerAlreadyRegisteredForTheMessageError
	}

	def, ok := rd.queueDefinitions[queue]
	if !ok {
		logrus.Error("bunmq queue definition not found for the given queue")
		return QueueDefinitionNotFoundError
	}

	rd.consumersDefinition[msgType] = &ConsumerDefinition{
		typ:             typ,
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

// ConsumeBlocking starts consuming messages from all registered queues
func (rd *dispatcher) ConsumeBlocking() {
	rd.mu.Lock()
	rd.consuming = true
	rd.mu.Unlock()

	logrus.Info("bunmq resilient dispatcher started, waiting for messages...")

	// Start all consumers
	for _, cd := range rd.consumersDefinition {
		go rd.consume(cd.typ, cd.queue, cd.msgType, cd.exchange, cd.routingKey)
	}

	// Wait for shutdown signal
	<-rd.signalCh
	logrus.Info("bunmq signal received, closing resilient dispatcher")

	rd.mu.Lock()
	rd.consuming = false
	rd.mu.Unlock()
}

// consume starts consuming messages from a specific queue with automatic reconnection
func (rd *dispatcher) consume(typ ConsumerDefinitionType, queue, msgType, exchange, routingKey string) {
	for {
		// Check if we should stop consuming
		rd.mu.RLock()
		if !rd.consuming {
			rd.mu.RUnlock()
			return
		}
		rd.mu.RUnlock()

		// Get current channel from connection manager
		ch, err := rd.connectionManager.GetChannel()
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

		rd.mu.Lock()
		rd.consumeChannels[queue] = delivery
		rd.mu.Unlock()

		logrus.Infof("bunmq started consuming from queue: %s", queue)

		// Process messages from this consumer
		for received := range delivery {
			// Process the message using the same logic as the original dispatcher
			rd.processMessage(typ, &received)
		}

		// If we reach here, the delivery channel was closed
		logrus.Warnf("bunmq delivery channel closed for queue: %s, will attempt to reconnect", queue)

		rd.mu.Lock()
		delete(rd.consumeChannels, queue)
		rd.mu.Unlock()

		// Wait a bit before trying to reconnect
		time.Sleep(time.Second * 2)
	}
}

// processMessage processes a single message (extracted from original dispatcher logic)
func (rd *dispatcher) processMessage(typ ConsumerDefinitionType, received *amqp.Delivery) {
	metadata, err := rd.extractMetadata(received)
	if err != nil {
		_ = received.Ack(false)
		return
	}

	def, err := rd.extractDefByType(typ, metadata)
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

	ctx, span := NewConsumerSpan(rd.tracer, received.Headers, received.Type)

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
			Warnf("bunmq message reprocessed too many times, sending to dead letter")
		_ = received.Ack(false)

		if err = rd.publishToDlq(def, received); err != nil {
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
				Warn("bunmq send message to process later")

			_ = received.Nack(false, false)
			span.End()
			return
		}

		if def.queueDefinition.withDLQ {
			span.RecordError(bunErr)
			_ = received.Ack(false)

			if err = rd.publishToDlq(def, received); err != nil {
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

// Helper methods (same as original dispatcher)
func (rd *dispatcher) extractMetadata(delivery *amqp.Delivery) (*DeliveryMetadata, error) {
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

func (rd *dispatcher) extractDefByType(typ ConsumerDefinitionType, metadata *DeliveryMetadata) (*ConsumerDefinition, error) {
	switch typ {
	case ConsumerDefinitionByType:
		def, ok := rd.consumersDefinition[metadata.Type]
		if !ok {
			return nil, fmt.Errorf("bunmq no consumer found for message type: %s", metadata.Type)
		}
		return def, nil
	case ConsumerDefinitionByExchange:
		for _, def := range rd.consumersDefinition {
			if def.exchange == metadata.OriginExchange {
				return def, nil
			}
		}
	case ConsumerDefinitionByRoutingKey:
		for _, def := range rd.consumersDefinition {
			if def.routingKey == metadata.RoutingKey {
				return def, nil
			}
		}
	case ConsumerDefinitionByExchangeRoutingKey:
		for _, def := range rd.consumersDefinition {
			if def.exchange == metadata.OriginExchange && def.routingKey == metadata.RoutingKey {
				return def, nil
			}
		}
	default:
		return nil, fmt.Errorf("bunmq unknown consumer definition type: %d", typ)
	}

	return nil, fmt.Errorf("bunmq no consumer definition found for type: %d", typ)
}

func (rd *dispatcher) publishToDlq(definition *ConsumerDefinition, received *amqp.Delivery) error {
	ch, err := rd.connectionManager.GetChannel()
	if err != nil {
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

// IsHealthy checks if the underlying connections are healthy
func (rd *dispatcher) IsHealthy() bool {
	return rd.connectionManager.IsHealthy()
}

// GetConnectionManager returns the underlying connection manager
func (rd *dispatcher) GetConnectionManager() ConnectionManager {
	return rd.connectionManager
}

// SetReconnectCallback sets a callback for when reconnection occurs
func (rd *dispatcher) SetReconnectCallback(callback func()) {
	rd.mu.Lock()
	defer rd.mu.Unlock()
	rd.reconnectCallback = callback
}
