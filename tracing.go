// Copyright (c) 2025 The GoKit Authors
// MIT License
// All rights reserved.

// Package amqp provides utilities for propagating trace context through AMQP (Advanced Message
// Queuing Protocol) messages. This enables distributed tracing across message-based
// microservices, allowing observability of end-to-end transactions across
// synchronous HTTP calls and asynchronous messaging.
//
// The package implements OpenTelemetry's TextMapCarrier interface for AMQP headers
// and provides functions to extract and inject trace context from/to AMQP messages.
package bunmq

import (
	"context"
	"fmt"
	"sort"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Traceparent represents the components of a trace context that are propagated between services.
// It contains the identifiers and flags needed to correlate spans across service boundaries
// when using message queuing.
type Traceparent struct {
	// TraceID is the globally unique identifier for a trace
	TraceID trace.TraceID

	// SpanID is the identifier for a specific operation within a trace
	SpanID trace.SpanID

	// TraceFlags contain options such as the sampling decision
	TraceFlags trace.TraceFlags
}

var (
	// AMQPPropagator is a composite propagator that combines TraceContext and Baggage propagation
	// for AMQP messaging contexts. This enables both trace correlation and contextual properties
	// to be passed between services.
	AMQPPropagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
)

// AMQPHeader wraps amqp.Table to implement the TextMapCarrier interface for OpenTelemetry propagation.
// This allows trace context to be injected into and extracted from AMQP message headers.
type AMQPHeader amqp.Table

// Set sets the value for the given key in the AMQP header (case-insensitive).
// This implements part of the TextMapCarrier interface required by OpenTelemetry
// for context propagation.
//
// Parameters:
//   - key: The header key (will be converted to lowercase)
//   - val: The header value to set
func (h AMQPHeader) Set(key, val string) {
	key = strings.ToLower(key)

	h[key] = val
}

// Get retrieves the value for a given key from the AMQP header (case-insensitive).
// This implements part of the TextMapCarrier interface required by OpenTelemetry
// for context propagation.
//
// Parameters:
//   - key: The header key to retrieve (will be converted to lowercase)
//
// Returns:
//   - string: The header value, or empty string if not found or not a string
func (h AMQPHeader) Get(key string) string {
	key = strings.ToLower(key)

	value, ok := h[key]

	if !ok {
		return ""
	}

	toString, ok := value.(string)

	if !ok {
		return ""
	}

	return toString
}

// Keys returns a sorted list of all keys in the AMQP header.
// This implements part of the TextMapCarrier interface required by OpenTelemetry
// for context propagation.
//
// Returns:
//   - []string: A sorted slice of all header keys
func (h AMQPHeader) Keys() []string {
	keys := make([]string, 0, len(h))

	for k := range h {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// NewConsumerSpan creates a new span for AMQP message consumption with the trace context
// extracted from the message headers. This allows continuation of a trace that was
// started in a producer service, maintaining the end-to-end transaction context.
//
// Parameters:
//   - tracer: The OpenTelemetry tracer to create the span
//   - header: The AMQP message headers containing the trace context
//   - typ: The type of consumer, used to name the span (e.g., queue name)
//
// Returns:
//   - context.Context: Context with the extracted trace information
//   - trace.Span: The new span created for this consumer operation
func NewConsumerSpan(tracer trace.Tracer, header amqp.Table, typ string) (context.Context, trace.Span) {
	ctx := AMQPPropagator.Extract(context.Background(), AMQPHeader(header))
	return tracer.Start(ctx, fmt.Sprintf("consume.%s", typ))
}
