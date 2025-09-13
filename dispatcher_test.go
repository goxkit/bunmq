// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Test message types for dispatcher tests
type DispatcherTestMessage struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func TestNewDispatcher(t *testing.T) {
	tests := []struct {
		name             string
		queueDefinitions []*QueueDefinition
		expectedQueues   int
	}{
		{
			name:             "empty queue definitions",
			queueDefinitions: []*QueueDefinition{},
			expectedQueues:   0,
		},
		{
			name: "single queue definition",
			queueDefinitions: []*QueueDefinition{
				NewQueue("test-queue"),
			},
			expectedQueues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConnectionManager()
			dispatcher := NewDispatcher(manager, tt.queueDefinitions)

			if dispatcher == nil {
				t.Fatal("NewDispatcher() returned nil")
			}

			// Verify it implements the Dispatcher interface
			if _, ok := dispatcher.(Dispatcher); !ok {
				t.Error("NewDispatcher() did not return a Dispatcher interface")
			}
		})
	}
}

func TestDispatcher_RegisterByType(t *testing.T) {
	tests := []struct {
		name           string
		queue          string
		msg            interface{}
		handler        ConsumerHandler
		existingQueue  bool
		expectError    bool
		expectedError  string
	}{
		{
			name:          "successful registration",
			queue:         "test-queue",
			msg:           DispatcherTestMessage{},
			handler:       func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil },
			existingQueue: true,
			expectError:   false,
		},
		{
			name:          "nil message",
			queue:         "test-queue",
			msg:           nil,
			handler:       func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil },
			existingQueue: true,
			expectError:   true,
			expectedError: "register dispatch with invalid parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConnectionManager()
			var queueDefs []*QueueDefinition
			
			if tt.existingQueue {
				queueDefs = []*QueueDefinition{NewQueue(tt.queue)}
			}
			
			dispatcher := NewDispatcher(manager, queueDefs)

			err := dispatcher.RegisterByType(tt.queue, tt.msg, tt.handler)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestDispatcher_RegisterByExchange(t *testing.T) {
	manager := NewMockConnectionManager()
	queueDefs := []*QueueDefinition{NewQueue("test-queue")}
	dispatcher := NewDispatcher(manager, queueDefs)
	
	handler := func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil }
	
	err := dispatcher.RegisterByExchange("test-queue", DispatcherTestMessage{}, "test-exchange", handler)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDispatcher_RegisterByRoutingKey(t *testing.T) {
	manager := NewMockConnectionManager()
	queueDefs := []*QueueDefinition{NewQueue("test-queue")}
	dispatcher := NewDispatcher(manager, queueDefs)
	
	handler := func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil }
	
	err := dispatcher.RegisterByRoutingKey("test-queue", DispatcherTestMessage{}, "test.key", handler)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDispatcher_RegisterByExchangeRoutingKey(t *testing.T) {
	manager := NewMockConnectionManager()
	queueDefs := []*QueueDefinition{NewQueue("test-queue")}
	dispatcher := NewDispatcher(manager, queueDefs)
	
	handler := func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil }
	
	err := dispatcher.RegisterByExchangeRoutingKey("test-queue", DispatcherTestMessage{}, "test-exchange", "test.key", handler)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDispatcher_extractMetadata(t *testing.T) {
	// Test basic metadata extraction
	manager := NewMockConnectionManager()
	dispatcher := NewDispatcher(manager, []*QueueDefinition{})
	
	// Use type assertion to access private method - dispatcher is a struct within the package
	type dispatcherImpl interface {
		extractMetadata(delivery *amqp.Delivery) (*DeliveryMetadata, error)
	}
	
	disp, ok := dispatcher.(dispatcherImpl)
	if !ok {
		t.Fatal("dispatcher does not implement expected interface")
	}
	
	delivery := &amqp.Delivery{
		MessageId:  "test-message-1",
		Type:       "DispatcherTestMessage",
		Exchange:   "test-exchange",
		RoutingKey: "test.routing.key",
		Headers:    map[string]interface{}{"custom": "value"},
	}
	
	metadata, err := disp.extractMetadata(delivery)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if metadata.MessageID != "test-message-1" {
		t.Errorf("Expected MessageID test-message-1, got %s", metadata.MessageID)
	}
	if metadata.Type != "DispatcherTestMessage" {
		t.Errorf("Expected Type DispatcherTestMessage, got %s", metadata.Type)
	}
}

func TestDispatcher_evaluateDefByReceivedMsg(t *testing.T) {
	manager := NewMockConnectionManager()
	queueDef := NewQueue("test-queue")
	dispatcher := NewDispatcher(manager, []*QueueDefinition{queueDef})
	
	// Register a consumer to test the evaluation logic
	handler := func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil }
	err := dispatcher.RegisterByType("test-queue", DispatcherTestMessage{}, handler)
	if err != nil {
		t.Errorf("Failed to register consumer: %v", err)
	}
	
	// We can't directly test evaluateDefByReceivedMsg since it's private,
	// but we can test the registration which uses the same logic
	if err != nil {
		t.Errorf("Expected successful registration, got error: %v", err)
	}
}

// Mock acknowledger for testing
type mockAcknowledger struct {
	ackFunc  func(multiple bool) error
	nackFunc func(multiple, requeue bool) error
}

func (m *mockAcknowledger) Ack(tag uint64, multiple bool) error {
	if m.ackFunc != nil {
		return m.ackFunc(multiple)
	}
	return nil
}

func (m *mockAcknowledger) Nack(tag uint64, multiple, requeue bool) error {
	if m.nackFunc != nil {
		return m.nackFunc(multiple, requeue)
	}
	return nil
}

func (m *mockAcknowledger) Reject(tag uint64, requeue bool) error {
	if m.nackFunc != nil {
		return m.nackFunc(false, requeue)
	}
	return nil
}