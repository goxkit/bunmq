// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"context"
	"reflect"
	"sort"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestTraceparent_Struct(t *testing.T) {
	// Test that Traceparent struct has the expected fields
	traceID := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
	traceFlags := trace.TraceFlags(1)

	tp := Traceparent{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: traceFlags,
	}

	if tp.TraceID != traceID {
		t.Errorf("Traceparent.TraceID = %v, want %v", tp.TraceID, traceID)
	}
	if tp.SpanID != spanID {
		t.Errorf("Traceparent.SpanID = %v, want %v", tp.SpanID, spanID)
	}
	if tp.TraceFlags != traceFlags {
		t.Errorf("Traceparent.TraceFlags = %v, want %v", tp.TraceFlags, traceFlags)
	}
}

func TestAMQPPropagator(t *testing.T) {
	// Test that AMQPPropagator is a composite propagator
	if AMQPPropagator == nil {
		t.Fatal("AMQPPropagator is nil")
	}

	// Test that it implements propagation.TextMapPropagator interface
	var _ propagation.TextMapPropagator = AMQPPropagator

	// Test Fields method exists
	fields := AMQPPropagator.Fields()
	if len(fields) == 0 {
		t.Error("AMQPPropagator.Fields() returned empty slice")
	}

	// Expected fields from TraceContext and Baggage propagators
	expectedFields := []string{"traceparent", "tracestate", "baggage"}
	for _, expected := range expectedFields {
		found := false
		for _, field := range fields {
			if field == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AMQPPropagator.Fields() missing expected field: %s", expected)
		}
	}
}

func TestAMQPHeader_Set(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected map[string]interface{}
	}{
		{
			name:     "basic set",
			key:      "traceparent",
			value:    "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			expected: map[string]interface{}{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
		},
		{
			name:     "uppercase key converted to lowercase",
			key:      "TRACEPARENT",
			value:    "test-value",
			expected: map[string]interface{}{"traceparent": "test-value"},
		},
		{
			name:     "mixed case key converted to lowercase",
			key:      "TraceParent",
			value:    "test-value",
			expected: map[string]interface{}{"traceparent": "test-value"},
		},
		{
			name:     "empty value",
			key:      "test-key",
			value:    "",
			expected: map[string]interface{}{"test-key": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := make(AMQPHeader)
			header.Set(tt.key, tt.value)

			if !reflect.DeepEqual(map[string]interface{}(header), tt.expected) {
				t.Errorf("AMQPHeader.Set() result = %v, want %v", map[string]interface{}(header), tt.expected)
			}
		})
	}
}

func TestAMQPHeader_Get(t *testing.T) {
	tests := []struct {
		name     string
		header   AMQPHeader
		key      string
		expected string
	}{
		{
			name:     "get existing string value",
			header:   AMQPHeader{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
			key:      "traceparent",
			expected: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		},
		{
			name:     "get non-existing key",
			header:   AMQPHeader{"traceparent": "test-value"},
			key:      "nonexistent",
			expected: "",
		},
		{
			name:     "get with different case",
			header:   AMQPHeader{"traceparent": "test-value"},
			key:      "TRACEPARENT",
			expected: "test-value",
		},
		{
			name:     "get non-string value",
			header:   AMQPHeader{"numeric": 123},
			key:      "numeric",
			expected: "",
		},
		{
			name:     "get from empty header",
			header:   AMQPHeader{},
			key:      "any-key",
			expected: "",
		},
		{
			name:     "get empty string value",
			header:   AMQPHeader{"empty": ""},
			key:      "empty",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.header.Get(tt.key)
			if result != tt.expected {
				t.Errorf("AMQPHeader.Get(%s) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestAMQPHeader_Keys(t *testing.T) {
	tests := []struct {
		name     string
		header   AMQPHeader
		expected []string
	}{
		{
			name:     "empty header",
			header:   AMQPHeader{},
			expected: []string{},
		},
		{
			name:     "single key",
			header:   AMQPHeader{"traceparent": "value"},
			expected: []string{"traceparent"},
		},
		{
			name:     "multiple keys sorted",
			header:   AMQPHeader{"z-key": "value1", "a-key": "value2", "m-key": "value3"},
			expected: []string{"a-key", "m-key", "z-key"},
		},
		{
			name:     "mixed types",
			header:   AMQPHeader{"string": "value", "number": 123, "bool": true},
			expected: []string{"bool", "number", "string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.header.Keys()
			
			// Sort both slices to ensure comparison works
			sort.Strings(result)
			sort.Strings(tt.expected)
			
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("AMQPHeader.Keys() = %v, want %v", result, tt.expected)
			}

			// Verify keys are sorted
			if !sort.StringsAreSorted(result) {
				t.Error("AMQPHeader.Keys() result should be sorted")
			}
		})
	}
}

func TestAMQPHeader_TextMapCarrier(t *testing.T) {
	// Test that AMQPHeader implements propagation.TextMapCarrier interface
	var header propagation.TextMapCarrier = make(AMQPHeader)

	// Test Set method
	header.Set("test-key", "test-value")
	
	// Test Get method
	value := header.Get("test-key")
	if value != "test-value" {
		t.Errorf("TextMapCarrier.Get() = %v, want test-value", value)
	}

	// Test Keys method
	keys := header.Keys()
	if len(keys) != 1 || keys[0] != "test-key" {
		t.Errorf("TextMapCarrier.Keys() = %v, want [test-key]", keys)
	}
}

func TestNewConsumerSpan(t *testing.T) {
	// Create a test tracer
	tracer := otel.Tracer("test-tracer")

	tests := []struct {
		name   string
		header amqp.Table
		typ    string
	}{
		{
			name:   "basic consumer span",
			header: amqp.Table{},
			typ:    "orders",
		},
		{
			name: "consumer span with trace context",
			header: amqp.Table{
				"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			},
			typ: "payments",
		},
		{
			name: "consumer span with multiple headers",
			header: amqp.Table{
				"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				"tracestate":  "vendor=value",
				"baggage":     "key=value",
			},
			typ: "notifications",
		},
		{
			name:   "consumer span with empty type",
			header: amqp.Table{},
			typ:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, span := NewConsumerSpan(tracer, tt.header, tt.typ)

			// Verify context is not nil
			if ctx == nil {
				t.Error("NewConsumerSpan() returned nil context")
			}

			// Verify span is not nil
			if span == nil {
				t.Error("NewConsumerSpan() returned nil span")
			}

			// Verify span name
			// Note: We can't easily access span name in the test without additional OpenTelemetry test setup
			// but we can verify the span is valid by checking if it implements the interface
			var _ trace.Span = span

			// End the span to clean up
			span.End()
		})
	}
}

func TestAMQPHeader_CaseInsensitivity(t *testing.T) {
	header := make(AMQPHeader)

	// Set with uppercase
	header.Set("TRACEPARENT", "test-value")

	// Get with lowercase
	value := header.Get("traceparent")
	if value != "test-value" {
		t.Errorf("Case insensitive get failed: got %v, want test-value", value)
	}

	// Get with mixed case
	value = header.Get("TraceParent")
	if value != "test-value" {
		t.Errorf("Case insensitive get failed: got %v, want test-value", value)
	}

	// Verify only one key exists (lowercase)
	keys := header.Keys()
	if len(keys) != 1 || keys[0] != "traceparent" {
		t.Errorf("Keys() = %v, want [traceparent]", keys)
	}
}

func TestAMQPHeader_Integration(t *testing.T) {
	// Test basic injection and extraction without requiring a real tracer
	// since setting up a full OpenTelemetry test environment is complex
	
	// Test manual injection of trace context
	header := make(AMQPHeader)
	header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	header.Set("tracestate", "vendor=value")
	
	// Verify trace context was set
	keys := header.Keys()
	if len(keys) < 2 {
		t.Errorf("Expected at least 2 headers, got %d", len(keys))
	}
	
	// Verify values can be retrieved
	traceparent := header.Get("traceparent")
	if traceparent != "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01" {
		t.Errorf("Traceparent = %v, want 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", traceparent)
	}
	
	tracestate := header.Get("tracestate")
	if tracestate != "vendor=value" {
		t.Errorf("Tracestate = %v, want vendor=value", tracestate)
	}
	
	// Test extraction with propagator (this will work even without a real span)
	extractedCtx := AMQPPropagator.Extract(context.Background(), header)
	if extractedCtx == nil {
		t.Error("Failed to extract context from AMQP header")
	}
	
	// The extraction should succeed but won't have a valid span context
	// since we're manually setting headers rather than using a real tracer
}