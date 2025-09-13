// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"testing"
	"time"
)

func TestNewQueue(t *testing.T) {
	tests := []struct {
		name         string
		queueName    string
		expectedName string
	}{
		{
			name:         "basic queue creation",
			queueName:    "test-queue",
			expectedName: "test-queue",
		},
		{
			name:         "empty queue name",
			queueName:    "",
			expectedName: "",
		},
		{
			name:         "queue with special characters",
			queueName:    "test-queue_123",
			expectedName: "test-queue_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQueue(tt.queueName)
			if q == nil {
				t.Fatal("NewQueue() returned nil")
			}
			if q.name != tt.expectedName {
				t.Errorf("NewQueue().name = %v, want %v", q.name, tt.expectedName)
			}
			// Test default values
			if !q.durable {
				t.Error("NewQueue() should create durable queue by default")
			}
			if q.delete {
				t.Error("NewQueue() should not create auto-delete queue by default")
			}
			if q.exclusive {
				t.Error("NewQueue() should not create exclusive queue by default")
			}
		})
	}
}

func TestQueueDefinition_Durable(t *testing.T) {
	q := NewQueue("test")

	// Test setting durable to true
	result := q.Durable(true)
	if result != q {
		t.Error("Durable() should return the same QueueDefinition instance")
	}
	if !q.durable {
		t.Error("Durable(true) should set durable to true")
	}

	// Test setting durable to false
	q.Durable(false)
	if q.durable {
		t.Error("Durable(false) should set durable to false")
	}
}

func TestQueueDefinition_Delete(t *testing.T) {
	q := NewQueue("test")

	// Test setting delete to true
	result := q.Delete(true)
	if result != q {
		t.Error("Delete() should return the same QueueDefinition instance")
	}
	if !q.delete {
		t.Error("Delete(true) should set delete to true")
	}

	// Test setting delete to false
	q.Delete(false)
	if q.delete {
		t.Error("Delete(false) should set delete to false")
	}
}

func TestQueueDefinition_Exclusive(t *testing.T) {
	q := NewQueue("test")

	// Test setting exclusive to true
	result := q.Exclusive(true)
	if result != q {
		t.Error("Exclusive() should return the same QueueDefinition instance")
	}
	if !q.exclusive {
		t.Error("Exclusive(true) should set exclusive to true")
	}

	// Test setting exclusive to false
	q.Exclusive(false)
	if q.exclusive {
		t.Error("Exclusive(false) should set exclusive to false")
	}
}

func TestQueueDefinition_WithMaxLength(t *testing.T) {
	q := NewQueue("test")

	tests := []struct {
		name     string
		length   int64
		expected int64
	}{
		{
			name:     "positive length",
			length:   1000,
			expected: 1000,
		},
		{
			name:     "zero length",
			length:   0,
			expected: 0,
		},
		{
			name:     "large length",
			length:   100000,
			expected: 100000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := q.WithMaxLength(tt.length)
			if result != q {
				t.Error("WithMaxLength() should return the same QueueDefinition instance")
			}
			if !q.withMaxLength {
				t.Error("WithMaxLength() should set withMaxLength to true")
			}
			if q.maxLength != tt.expected {
				t.Errorf("WithMaxLength(%d) should set maxLength to %d, got %d", tt.length, tt.expected, q.maxLength)
			}
		})
	}
}

func TestQueueDefinition_WithTTL(t *testing.T) {
	q := NewQueue("test")

	tests := []struct {
		name     string
		ttl      time.Duration
		expected time.Duration
	}{
		{
			name:     "seconds TTL",
			ttl:      30 * time.Second,
			expected: 30 * time.Second,
		},
		{
			name:     "minutes TTL",
			ttl:      5 * time.Minute,
			expected: 5 * time.Minute,
		},
		{
			name:     "zero TTL",
			ttl:      0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := q.WithTTL(tt.ttl)
			if result != q {
				t.Error("WithTTL() should return the same QueueDefinition instance")
			}
			if !q.withTTL {
				t.Error("WithTTL() should set withTTL to true")
			}
			if q.ttl != tt.expected {
				t.Errorf("WithTTL(%v) should set ttl to %v, got %v", tt.ttl, tt.expected, q.ttl)
			}
		})
	}
}

func TestQueueDefinition_WithDQL(t *testing.T) {
	tests := []struct {
		name          string
		queueName     string
		expectedDLQ   string
	}{
		{
			name:        "basic DLQ",
			queueName:   "orders",
			expectedDLQ: "orders-dlq",
		},
		{
			name:        "DLQ with dashes",
			queueName:   "order-processing",
			expectedDLQ: "order-processing-dlq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQueue(tt.queueName)
			result := q.WithDQL()
			if result != q {
				t.Error("WithDQL() should return the same QueueDefinition instance")
			}
			if !q.withDLQ {
				t.Error("WithDQL() should set withDLQ to true")
			}
			if q.dqlName != tt.expectedDLQ {
				t.Errorf("WithDQL() should set dqlName to %s, got %s", tt.expectedDLQ, q.dqlName)
			}
		})
	}
}

func TestQueueDefinition_WithDLQMaxLength(t *testing.T) {
	q := NewQueue("test")

	tests := []struct {
		name     string
		length   int64
		expected int64
	}{
		{
			name:     "positive DLQ length",
			length:   5000,
			expected: 5000,
		},
		{
			name:     "zero DLQ length",
			length:   0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := q.WithDLQMaxLength(tt.length)
			if result != q {
				t.Error("WithDLQMaxLength() should return the same QueueDefinition instance")
			}
			if !q.withDLQMaxLength {
				t.Error("WithDLQMaxLength() should set withDLQMaxLength to true")
			}
			if q.dlqMaxLength != tt.expected {
				t.Errorf("WithDLQMaxLength(%d) should set dlqMaxLength to %d, got %d", tt.length, tt.expected, q.dlqMaxLength)
			}
		})
	}
}

func TestQueueDefinition_WithRetry(t *testing.T) {
	q := NewQueue("test")

	tests := []struct {
		name            string
		ttl             time.Duration
		retries         int64
		expectedTTL     time.Duration
		expectedRetries int64
	}{
		{
			name:            "basic retry",
			ttl:             30 * time.Second,
			retries:         3,
			expectedTTL:     30 * time.Second,
			expectedRetries: 3,
		},
		{
			name:            "zero retries",
			ttl:             10 * time.Second,
			retries:         0,
			expectedTTL:     10 * time.Second,
			expectedRetries: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := q.WithRetry(tt.ttl, tt.retries)
			if result != q {
				t.Error("WithRetry() should return the same QueueDefinition instance")
			}
			if !q.withRetry {
				t.Error("WithRetry() should set withRetry to true")
			}
			if q.retryTTL != tt.expectedTTL {
				t.Errorf("WithRetry() should set retryTTL to %v, got %v", tt.expectedTTL, q.retryTTL)
			}
			if q.retries != tt.expectedRetries {
				t.Errorf("WithRetry() should set retries to %d, got %d", tt.expectedRetries, q.retries)
			}
		})
	}
}

func TestQueueDefinition_Name(t *testing.T) {
	tests := []struct {
		name      string
		queueName string
		expected  string
	}{
		{
			name:      "basic name",
			queueName: "orders",
			expected:  "orders",
		},
		{
			name:      "empty name",
			queueName: "",
			expected:  "",
		},
		{
			name:      "name with special chars",
			queueName: "order_queue-123",
			expected:  "order_queue-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQueue(tt.queueName)
			if q.Name() != tt.expected {
				t.Errorf("Name() = %v, want %v", q.Name(), tt.expected)
			}
		})
	}
}

func TestQueueDefinition_DLQName(t *testing.T) {
	tests := []struct {
		name      string
		queueName string
		expected  string
	}{
		{
			name:      "basic DLQ name",
			queueName: "orders",
			expected:  "orders-dlq",
		},
		{
			name:      "empty queue name",
			queueName: "",
			expected:  "-dlq",
		},
		{
			name:      "queue with dashes",
			queueName: "order-processing",
			expected:  "order-processing-dlq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQueue(tt.queueName)
			if q.DLQName() != tt.expected {
				t.Errorf("DLQName() = %v, want %v", q.DLQName(), tt.expected)
			}
		})
	}
}

func TestQueueDefinition_RetryName(t *testing.T) {
	tests := []struct {
		name      string
		queueName string
		expected  string
	}{
		{
			name:      "basic retry name",
			queueName: "orders",
			expected:  "orders-retry",
		},
		{
			name:      "empty queue name",
			queueName: "",
			expected:  "-retry",
		},
		{
			name:      "queue with dashes",
			queueName: "order-processing",
			expected:  "order-processing-retry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQueue(tt.queueName)
			if q.RetryName() != tt.expected {
				t.Errorf("RetryName() = %v, want %v", q.RetryName(), tt.expected)
			}
		})
	}
}

func TestQueueDefinition_FluentChaining(t *testing.T) {
	// Test that all methods can be chained together
	q := NewQueue("test-queue").
		Durable(true).
		Delete(false).
		Exclusive(false).
		WithMaxLength(10000).
		WithTTL(5 * time.Minute).
		WithDQL().
		WithDLQMaxLength(1000).
		WithRetry(30*time.Second, 3)

	if q.name != "test-queue" {
		t.Errorf("Chained queue name = %v, want test-queue", q.name)
	}
	if !q.durable {
		t.Error("Chained queue should be durable")
	}
	if q.delete {
		t.Error("Chained queue should not be auto-delete")
	}
	if q.exclusive {
		t.Error("Chained queue should not be exclusive")
	}
	if !q.withMaxLength || q.maxLength != 10000 {
		t.Error("Chained queue should have max length of 10000")
	}
	if !q.withTTL || q.ttl != 5*time.Minute {
		t.Error("Chained queue should have TTL of 5 minutes")
	}
	if !q.withDLQ || q.dqlName != "test-queue-dlq" {
		t.Error("Chained queue should have DLQ enabled")
	}
	if !q.withDLQMaxLength || q.dlqMaxLength != 1000 {
		t.Error("Chained queue should have DLQ max length of 1000")
	}
	if !q.withRetry || q.retryTTL != 30*time.Second || q.retries != 3 {
		t.Error("Chained queue should have retry enabled with 30s TTL and 3 retries")
	}
}