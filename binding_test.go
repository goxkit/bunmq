// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"testing"
)

func TestNewExchangeBiding(t *testing.T) {
	binding := NewExchangeBiding()
	if binding == nil {
		t.Fatal("NewExchangeBiding() returned nil")
	}

	// Test zero values
	if binding.source != "" {
		t.Errorf("NewExchangeBiding().source = %v, want empty string", binding.source)
	}
	if binding.destination != "" {
		t.Errorf("NewExchangeBiding().destination = %v, want empty string", binding.destination)
	}
	if binding.routingKey != "" {
		t.Errorf("NewExchangeBiding().routingKey = %v, want empty string", binding.routingKey)
	}
	if binding.args != nil {
		t.Errorf("NewExchangeBiding().args = %v, want nil", binding.args)
	}
}

func TestNewQueueBinding(t *testing.T) {
	binding := NewQueueBinding()
	if binding == nil {
		t.Fatal("NewQueueBinding() returned nil")
	}

	// Test zero values
	if binding.routingKey != "" {
		t.Errorf("NewQueueBinding().routingKey = %v, want empty string", binding.routingKey)
	}
	if binding.queue != "" {
		t.Errorf("NewQueueBinding().queue = %v, want empty string", binding.queue)
	}
	if binding.exchange != "" {
		t.Errorf("NewQueueBinding().exchange = %v, want empty string", binding.exchange)
	}
	if binding.args != nil {
		t.Errorf("NewQueueBinding().args = %v, want nil", binding.args)
	}
}

func TestQueueBindingDefinition_RoutingKey(t *testing.T) {
	binding := NewQueueBinding()

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "basic routing key",
			key:      "order.created",
			expected: "order.created",
		},
		{
			name:     "empty routing key",
			key:      "",
			expected: "",
		},
		{
			name:     "complex routing key",
			key:      "user.account.updated",
			expected: "user.account.updated",
		},
		{
			name:     "wildcard routing key",
			key:      "*.order.*",
			expected: "*.order.*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := binding.RoutingKey(tt.key)
			if result != binding {
				t.Error("RoutingKey() should return the same QueueBindingDefinition instance")
			}
			if binding.routingKey != tt.expected {
				t.Errorf("RoutingKey(%s) set routingKey to %v, want %v", tt.key, binding.routingKey, tt.expected)
			}
		})
	}
}

func TestQueueBindingDefinition_Queue(t *testing.T) {
	binding := NewQueueBinding()

	tests := []struct {
		name     string
		queue    string
		expected string
	}{
		{
			name:     "basic queue name",
			queue:    "orders",
			expected: "orders",
		},
		{
			name:     "empty queue name",
			queue:    "",
			expected: "",
		},
		{
			name:     "queue with special chars",
			queue:    "order_queue-123",
			expected: "order_queue-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := binding.Queue(tt.queue)
			if result != binding {
				t.Error("Queue() should return the same QueueBindingDefinition instance")
			}
			if binding.queue != tt.expected {
				t.Errorf("Queue(%s) set queue to %v, want %v", tt.queue, binding.queue, tt.expected)
			}
		})
	}
}

func TestQueueBindingDefinition_Exchange(t *testing.T) {
	binding := NewQueueBinding()

	tests := []struct {
		name     string
		exchange string
		expected string
	}{
		{
			name:     "basic exchange name",
			exchange: "orders",
			expected: "orders",
		},
		{
			name:     "empty exchange name",
			exchange: "",
			expected: "",
		},
		{
			name:     "exchange with special chars",
			exchange: "order_exchange-123",
			expected: "order_exchange-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := binding.Exchange(tt.exchange)
			if result != binding {
				t.Error("Exchange() should return the same QueueBindingDefinition instance")
			}
			if binding.exchange != tt.expected {
				t.Errorf("Exchange(%s) set exchange to %v, want %v", tt.exchange, binding.exchange, tt.expected)
			}
		})
	}
}

func TestQueueBindingDefinition_Args(t *testing.T) {
	binding := NewQueueBinding()

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "basic arguments",
			args:     map[string]interface{}{"x-match": "all", "format": "json"},
			expected: map[string]interface{}{"x-match": "all", "format": "json"},
		},
		{
			name:     "empty arguments",
			args:     map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:     "nil arguments",
			args:     nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := binding.Args(tt.args)
			if result != binding {
				t.Error("Args() should return the same QueueBindingDefinition instance")
			}

			if tt.args == nil && binding.args != nil {
				t.Error("Args(nil) should set args to nil")
			} else if tt.args != nil {
				if len(binding.args) != len(tt.expected) {
					t.Errorf("Args() set %d arguments, want %d", len(binding.args), len(tt.expected))
				}
				for key, expectedValue := range tt.expected {
					if actualValue, exists := binding.args[key]; !exists {
						t.Errorf("Args() missing key %s", key)
					} else if actualValue != expectedValue {
						t.Errorf("Args() key %s = %v, want %v", key, actualValue, expectedValue)
					}
				}
			}
		})
	}
}

func TestExchangeBindingDefinition_Source(t *testing.T) {
	binding := NewExchangeBiding()

	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:     "basic source name",
			source:   "orders",
			expected: "orders",
		},
		{
			name:     "empty source name",
			source:   "",
			expected: "",
		},
		{
			name:     "source with special chars",
			source:   "order_exchange-123",
			expected: "order_exchange-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := binding.Source(tt.source)
			if result != binding {
				t.Error("Source() should return the same ExchangeBindingDefinition instance")
			}
			if binding.source != tt.expected {
				t.Errorf("Source(%s) set source to %v, want %v", tt.source, binding.source, tt.expected)
			}
		})
	}
}

func TestExchangeBindingDefinition_Destination(t *testing.T) {
	binding := NewExchangeBiding()

	tests := []struct {
		name        string
		destination string
		expected    string
	}{
		{
			name:        "basic destination name",
			destination: "notifications",
			expected:    "notifications",
		},
		{
			name:        "empty destination name",
			destination: "",
			expected:    "",
		},
		{
			name:        "destination with special chars",
			destination: "notification_exchange-123",
			expected:    "notification_exchange-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := binding.Destination(tt.destination)
			if result != binding {
				t.Error("Destination() should return the same ExchangeBindingDefinition instance")
			}
			if binding.destination != tt.expected {
				t.Errorf("Destination(%s) set destination to %v, want %v", tt.destination, binding.destination, tt.expected)
			}
		})
	}
}

func TestExchangeBindingDefinition_RoutingKey(t *testing.T) {
	binding := NewExchangeBiding()

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "basic routing key",
			key:      "order.created",
			expected: "order.created",
		},
		{
			name:     "empty routing key",
			key:      "",
			expected: "",
		},
		{
			name:     "complex routing key",
			key:      "user.account.updated",
			expected: "user.account.updated",
		},
		{
			name:     "wildcard routing key",
			key:      "*.order.*",
			expected: "*.order.*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := binding.RoutingKey(tt.key)
			if result != binding {
				t.Error("RoutingKey() should return the same ExchangeBindingDefinition instance")
			}
			if binding.routingKey != tt.expected {
				t.Errorf("RoutingKey(%s) set routingKey to %v, want %v", tt.key, binding.routingKey, tt.expected)
			}
		})
	}
}

func TestExchangeBindingDefinition_Args(t *testing.T) {
	binding := NewExchangeBiding()

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "basic arguments",
			args:     map[string]interface{}{"x-match": "any", "content-type": "application/json"},
			expected: map[string]interface{}{"x-match": "any", "content-type": "application/json"},
		},
		{
			name:     "empty arguments",
			args:     map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:     "nil arguments",
			args:     nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := binding.Args(tt.args)
			if result != binding {
				t.Error("Args() should return the same ExchangeBindingDefinition instance")
			}

			if tt.args == nil && binding.args != nil {
				t.Error("Args(nil) should set args to nil")
			} else if tt.args != nil {
				if len(binding.args) != len(tt.expected) {
					t.Errorf("Args() set %d arguments, want %d", len(binding.args), len(tt.expected))
				}
				for key, expectedValue := range tt.expected {
					if actualValue, exists := binding.args[key]; !exists {
						t.Errorf("Args() missing key %s", key)
					} else if actualValue != expectedValue {
						t.Errorf("Args() key %s = %v, want %v", key, actualValue, expectedValue)
					}
				}
			}
		})
	}
}

func TestQueueBindingDefinition_FluentChaining(t *testing.T) {
	// Test that all methods can be chained together
	args := map[string]interface{}{"x-match": "all"}
	binding := NewQueueBinding().
		RoutingKey("order.created").
		Queue("orders").
		Exchange("order-exchange").
		Args(args)

	if binding.routingKey != "order.created" {
		t.Errorf("Chained binding routingKey = %v, want order.created", binding.routingKey)
	}
	if binding.queue != "orders" {
		t.Errorf("Chained binding queue = %v, want orders", binding.queue)
	}
	if binding.exchange != "order-exchange" {
		t.Errorf("Chained binding exchange = %v, want order-exchange", binding.exchange)
	}
	if len(binding.args) != 1 || binding.args["x-match"] != "all" {
		t.Error("Chained binding should have the correct arguments")
	}
}

func TestExchangeBindingDefinition_FluentChaining(t *testing.T) {
	// Test that all methods can be chained together
	args := map[string]interface{}{"alternate-exchange": "backup"}
	binding := NewExchangeBiding().
		Source("orders").
		Destination("notifications").
		RoutingKey("order.created").
		Args(args)

	if binding.source != "orders" {
		t.Errorf("Chained binding source = %v, want orders", binding.source)
	}
	if binding.destination != "notifications" {
		t.Errorf("Chained binding destination = %v, want notifications", binding.destination)
	}
	if binding.routingKey != "order.created" {
		t.Errorf("Chained binding routingKey = %v, want order.created", binding.routingKey)
	}
	if len(binding.args) != 1 || binding.args["alternate-exchange"] != "backup" {
		t.Error("Chained binding should have the correct arguments")
	}
}