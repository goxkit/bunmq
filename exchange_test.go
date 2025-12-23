// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"testing"
)

func TestExchangeKind_String(t *testing.T) {
	tests := []struct {
		name     string
		kind     ExchangeKind
		expected string
	}{
		{
			name:     "FanoutExchange",
			kind:     FanoutExchange,
			expected: "fanout",
		},
		{
			name:     "DirectExchange",
			kind:     DirectExchange,
			expected: "direct",
		},
		{
			name:     "TopicExchange",
			kind:     TopicExchange,
			expected: "topic",
		},
		{
			name:     "HeadersExchange",
			kind:     HeadersExchange,
			expected: "headers",
		},
		{
			name:     "XDelayedMessageExchange",
			kind:     XDelayedMessageExchange,
			expected: "x-delayed-message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.expected {
				t.Errorf("ExchangeKind.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExchangeKind_Constants(t *testing.T) {
	// Test that exchange kind constants have correct values
	if FanoutExchange != "fanout" {
		t.Errorf("FanoutExchange = %v, want fanout", FanoutExchange)
	}
	if DirectExchange != "direct" {
		t.Errorf("DirectExchange = %v, want direct", DirectExchange)
	}
	if TopicExchange != "topic" {
		t.Errorf("TopicExchange = %v, want topic", TopicExchange)
	}
	if HeadersExchange != "headers" {
		t.Errorf("HeadersExchange = %v, want headers", HeadersExchange)
	}
	if XDelayedMessageExchange != "x-delayed-message" {
		t.Errorf("XDelayedMessageExchange = %v, want x-delayed-message", XDelayedMessageExchange)
	}
}

func TestNewDirectExchange(t *testing.T) {
	tests := []struct {
		name         string
		exchangeName string
		expected     *ExchangeDefinition
	}{
		{
			name:         "basic direct exchange",
			exchangeName: "orders",
			expected: &ExchangeDefinition{
				name:    "orders",
				durable: true,
				delete:  false,
				kind:    DirectExchange,
			},
		},
		{
			name:         "empty name",
			exchangeName: "",
			expected: &ExchangeDefinition{
				name:    "",
				durable: true,
				delete:  false,
				kind:    DirectExchange,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchange := NewDirectExchange(tt.exchangeName)
			if exchange == nil {
				t.Fatal("NewDirectExchange() returned nil")
				return
			}
			if exchange.name != tt.expected.name {
				t.Errorf("NewDirectExchange().name = %v, want %v", exchange.name, tt.expected.name)
			}
			if exchange.durable != tt.expected.durable {
				t.Errorf("NewDirectExchange().durable = %v, want %v", exchange.durable, tt.expected.durable)
			}
			if exchange.delete != tt.expected.delete {
				t.Errorf("NewDirectExchange().delete = %v, want %v", exchange.delete, tt.expected.delete)
			}
			if exchange.kind != tt.expected.kind {
				t.Errorf("NewDirectExchange().kind = %v, want %v", exchange.kind, tt.expected.kind)
			}
		})
	}
}

func TestNewFanoutExchange(t *testing.T) {
	tests := []struct {
		name         string
		exchangeName string
		expected     *ExchangeDefinition
	}{
		{
			name:         "basic fanout exchange",
			exchangeName: "notifications",
			expected: &ExchangeDefinition{
				name:    "notifications",
				durable: true,
				delete:  false,
				kind:    FanoutExchange,
			},
		},
		{
			name:         "empty name",
			exchangeName: "",
			expected: &ExchangeDefinition{
				name:    "",
				durable: true,
				delete:  false,
				kind:    FanoutExchange,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchange := NewFanoutExchange(tt.exchangeName)
			if exchange == nil {
				t.Fatal("NewFanoutExchange() returned nil")
				return
			}
			if exchange.name != tt.expected.name {
				t.Errorf("NewFanoutExchange().name = %v, want %v", exchange.name, tt.expected.name)
			}
			if exchange.durable != tt.expected.durable {
				t.Errorf("NewFanoutExchange().durable = %v, want %v", exchange.durable, tt.expected.durable)
			}
			if exchange.delete != tt.expected.delete {
				t.Errorf("NewFanoutExchange().delete = %v, want %v", exchange.delete, tt.expected.delete)
			}
			if exchange.kind != tt.expected.kind {
				t.Errorf("NewFanoutExchange().kind = %v, want %v", exchange.kind, tt.expected.kind)
			}
		})
	}
}

func TestNewDirectExchanges(t *testing.T) {
	tests := []struct {
		name     string
		names    []string
		expected int
	}{
		{
			name:     "multiple exchanges",
			names:    []string{"orders", "payments", "inventory"},
			expected: 3,
		},
		{
			name:     "single exchange",
			names:    []string{"orders"},
			expected: 1,
		},
		{
			name:     "empty slice",
			names:    []string{},
			expected: 0,
		},
		{
			name:     "nil slice",
			names:    nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchanges := NewDirectExchanges(tt.names)
			if len(exchanges) != tt.expected {
				t.Errorf("NewDirectExchanges() returned %d exchanges, want %d", len(exchanges), tt.expected)
			}

			for i, exchange := range exchanges {
				if exchange == nil {
					t.Errorf("NewDirectExchanges()[%d] is nil", i)
					continue
				}
				if i < len(tt.names) && exchange.name != tt.names[i] {
					t.Errorf("NewDirectExchanges()[%d].name = %v, want %v", i, exchange.name, tt.names[i])
				}
				if exchange.kind != DirectExchange {
					t.Errorf("NewDirectExchanges()[%d].kind = %v, want %v", i, exchange.kind, DirectExchange)
				}
				if !exchange.durable {
					t.Errorf("NewDirectExchanges()[%d].durable = false, want true", i)
				}
				if exchange.delete {
					t.Errorf("NewDirectExchanges()[%d].delete = true, want false", i)
				}
			}
		})
	}
}

func TestNewFanoutExchanges(t *testing.T) {
	tests := []struct {
		name     string
		names    []string
		expected int
	}{
		{
			name:     "multiple exchanges",
			names:    []string{"notifications", "broadcasts", "alerts"},
			expected: 3,
		},
		{
			name:     "single exchange",
			names:    []string{"notifications"},
			expected: 1,
		},
		{
			name:     "empty slice",
			names:    []string{},
			expected: 0,
		},
		{
			name:     "nil slice",
			names:    nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchanges := NewFanoutExchanges(tt.names)
			if len(exchanges) != tt.expected {
				t.Errorf("NewFanoutExchanges() returned %d exchanges, want %d", len(exchanges), tt.expected)
			}

			for i, exchange := range exchanges {
				if exchange == nil {
					t.Errorf("NewFanoutExchanges()[%d] is nil", i)
					continue
				}
				if i < len(tt.names) && exchange.name != tt.names[i] {
					t.Errorf("NewFanoutExchanges()[%d].name = %v, want %v", i, exchange.name, tt.names[i])
				}
				if exchange.kind != FanoutExchange {
					t.Errorf("NewFanoutExchanges()[%d].kind = %v, want %v", i, exchange.kind, FanoutExchange)
				}
				if !exchange.durable {
					t.Errorf("NewFanoutExchanges()[%d].durable = false, want true", i)
				}
				if exchange.delete {
					t.Errorf("NewFanoutExchanges()[%d].delete = true, want false", i)
				}
			}
		})
	}
}

func TestExchangeDefinition_Durable(t *testing.T) {
	exchange := NewDirectExchange("test")

	// Test setting durable to true
	result := exchange.Durable(true)
	if result != exchange {
		t.Error("Durable() should return the same ExchangeDefinition instance")
	}
	if !exchange.durable {
		t.Error("Durable(true) should set durable to true")
	}

	// Test setting durable to false
	exchange.Durable(false)
	if exchange.durable {
		t.Error("Durable(false) should set durable to false")
	}
}

func TestExchangeDefinition_Delete(t *testing.T) {
	exchange := NewDirectExchange("test")

	// Test setting delete to true
	result := exchange.Delete(true)
	if result != exchange {
		t.Error("Delete() should return the same ExchangeDefinition instance")
	}
	if !exchange.delete {
		t.Error("Delete(true) should set delete to true")
	}

	// Test setting delete to false
	exchange.Delete(false)
	if exchange.delete {
		t.Error("Delete(false) should set delete to false")
	}
}

func TestExchangeDefinition_Params(t *testing.T) {
	exchange := NewDirectExchange("test")

	tests := []struct {
		name     string
		params   map[string]any
		expected map[string]any
	}{
		{
			name:     "basic parameters",
			params:   map[string]any{"type": "direct", "priority": 1},
			expected: map[string]any{"type": "direct", "priority": 1},
		},
		{
			name:     "empty parameters",
			params:   map[string]any{},
			expected: map[string]any{},
		},
		{
			name:     "nil parameters",
			params:   nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exchange.Params(tt.params)
			if result != exchange {
				t.Error("Params() should return the same ExchangeDefinition instance")
			}

			if tt.params == nil && exchange.params != nil {
				t.Error("Params(nil) should set params to nil")
			} else if tt.params != nil {
				if len(exchange.params) != len(tt.expected) {
					t.Errorf("Params() set %d parameters, want %d", len(exchange.params), len(tt.expected))
				}
				for key, expectedValue := range tt.expected {
					if actualValue, exists := exchange.params[key]; !exists {
						t.Errorf("Params() missing key %s", key)
					} else if actualValue != expectedValue {
						t.Errorf("Params() key %s = %v, want %v", key, actualValue, expectedValue)
					}
				}
			}
		})
	}
}

func TestExchangeDefinition_FluentChaining(t *testing.T) {
	// Test that all methods can be chained together
	params := map[string]any{"alternate-exchange": "backup"}
	exchange := NewDirectExchange("test-exchange").
		Durable(false).
		Delete(true).
		Params(params)

	if exchange.name != "test-exchange" {
		t.Errorf("Chained exchange name = %v, want test-exchange", exchange.name)
	}
	if exchange.durable {
		t.Error("Chained exchange should not be durable")
	}
	if !exchange.delete {
		t.Error("Chained exchange should be auto-delete")
	}
	if exchange.kind != DirectExchange {
		t.Errorf("Chained exchange kind = %v, want %v", exchange.kind, DirectExchange)
	}
	if len(exchange.params) != 1 || exchange.params["alternate-exchange"] != "backup" {
		t.Error("Chained exchange should have the correct parameters")
	}
}

func TestDefaultExchange(t *testing.T) {
	tests := []struct {
		name         string
		exchangeName string
		kind         ExchangeKind
	}{
		{
			name:         "direct exchange",
			exchangeName: "test-direct",
			kind:         DirectExchange,
		},
		{
			name:         "fanout exchange",
			exchangeName: "test-fanout",
			kind:         FanoutExchange,
		},
		{
			name:         "topic exchange",
			exchangeName: "test-topic",
			kind:         TopicExchange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchange := defaultExchange(tt.exchangeName, tt.kind)
			if exchange == nil {
				t.Fatal("defaultExchange() returned nil")
				return
			}
			if exchange.name != tt.exchangeName {
				t.Errorf("defaultExchange().name = %v, want %v", exchange.name, tt.exchangeName)
			}
			if exchange.kind != tt.kind {
				t.Errorf("defaultExchange().kind = %v, want %v", exchange.kind, tt.kind)
			}
			if !exchange.durable {
				t.Error("defaultExchange() should create durable exchange")
			}
			if exchange.delete {
				t.Error("defaultExchange() should not create auto-delete exchange")
			}
		})
	}
}
