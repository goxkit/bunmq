// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type (
	// Publisher defines an interface for publishing messages to a messaging system.
	// It provides methods for sending messages with optional metadata such as
	// destination, source, and routing keys.
	Publisher interface {
		// Publish sends a message to the specified destination.
		//
		// Parameters:
		// - ctx: The context for managing deadlines, cancellations, and other request-scoped values.
		// - to: The destination or topic where the message should be sent (optional).
		// - from: The source or origin of the message (optional).
		// - key: A routing key or identifier for the message (optional).
		// - msg: The message payload to be sent.
		// - options: Additional dynamic parameters for the message (optional).
		//
		// Returns:
		// - An error if the message could not be sent.
		Publish(ctx context.Context, to, from, key *string, msg any, options ...*Option) error

		// PublishDeadline sends a message to the specified destination with a deadline.
		// This method ensures that the message is sent within the context's deadline.
		//
		// Parameters:
		// - ctx: The context for managing deadlines, cancellations, and other request-scoped values.
		// - to: The destination or topic where the message should be sent (optional).
		// - from: The source or origin of the message (optional).
		// - key: A routing key or identifier for the message (optional).
		// - msg: The message payload to be sent.
		// - options: Additional dynamic parameters for the message (optional).
		//
		// Returns:
		// - An error if the message could not be sent within the deadline.
		PublishDeadline(ctx context.Context, to, from, key *string, msg any, options ...*Option) error
	}

	// publisher is the concrete implementation of the Publisher interface.
	// It handles the details of marshaling messages, setting headers, and publishing to RabbitMQ.
	publisher struct {
		appName string
		channel AMQPChannel
	}
)

// JsonContentType is the MIME type used for JSON message content.
const (
	JsonContentType = "application/json"
)

// NewPublisher creates a new publisher instance with the provided configuration and AMQP channel.
func NewPublisher(appName string, channel AMQPChannel) Publisher {
	return &publisher{appName, channel}
}

// SimplePublish publishes a message directly to a target queue.
// The exchange is left empty, which means the default exchange is used.
func (p *publisher) SimplePublish(ctx context.Context, target string, msg any) error {
	return p.publish(ctx, target, "", msg)
}

// Publish publishes a message to a specified exchange with optional routing key.
// It aligns with the Publisher interface and handles tracing propagation.
// Parameters:
//   - ctx: Context for tracing and cancellation
//   - to: Pointer to the target exchange name (required)
//   - from: Pointer to source identifier (optional, not used)
//   - key: Pointer to routing key (optional)
//   - msg: The message to publish (will be marshaled to JSON)
//   - options: Additional publishing options (optional)
//
// Returns an error if publishing fails or if the exchange name is empty.
func (p *publisher) Publish(ctx context.Context, to, from, key *string, msg any, options ...*Option) error {
	if to == nil || *to == "" {
		logrus.WithContext(ctx).Error("exchange cannot be empty")
		return fmt.Errorf("exchange cannot be empty")
	}

	exchange := *to
	routingKey := ""
	if key != nil {
		routingKey = *key
	}

	return p.publish(ctx, exchange, routingKey, msg)
}

// PublishDeadline publishes a message to a specified exchange with a deadline.
// It's similar to Publish but with an added timeout of 1 second.
// Parameters:
//   - ctx: Context for tracing and cancellation
//   - to: Pointer to the target exchange name (required)
//   - from: Pointer to source identifier (optional, not used)
//   - key: Pointer to routing key (optional)
//   - msg: The message to publish (will be marshaled to JSON)
//   - options: Additional publishing options (optional)
//
// Returns an error if publishing fails, times out, or if the exchange name is empty.
func (p *publisher) PublishDeadline(ctx context.Context, to, from, key *string, msg any, options ...*Option) error {
	if to == nil || *to == "" {
		logrus.WithContext(ctx).Error("exchange cannot be empty")
		return fmt.Errorf("exchange cannot be empty")
	}

	exchange := *to
	routingKey := ""
	if key != nil {
		routingKey = *key
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	return p.publish(ctx, exchange, routingKey, msg)
}

// publish is the internal method that handles the details of publishing a message.
// It marshals the message to JSON, sets headers for tracing, and publishes to RabbitMQ.
func (p *publisher) publish(ctx context.Context, exchange, key string, msg any) error {
	byt, err := json.Marshal(msg)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("publisher marshal")
		return err
	}

	headers := amqp.Table{}
	AMQPPropagator.Inject(ctx, AMQPHeader(headers))

	mID, err := uuid.NewV7()
	if err != nil {
		mID = uuid.New()
	}

	return p.channel.Publish(exchange, key, false, false, amqp.Publishing{
		Headers:     headers,
		Type:        fmt.Sprintf("%T", msg),
		ContentType: JsonContentType,
		MessageId:   mID.String(),
		AppId:       p.appName,
		Body:        byt,
	})
}
