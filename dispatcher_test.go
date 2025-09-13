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

			// NewDispatcher returns a Dispatcher interface; no runtime type assertion is needed here
			// since dispatcher is already declared with that interface type.
		})
	}
}

func TestDispatcher_RegisterByType(t *testing.T) {
	tests := []struct {
		name          string
		queue         string
		msg           interface{}
		handler       ConsumerHandler
		existingQueue bool
		expectError   bool
		expectedError string
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

func TestDispatcher_ConsumeBlocking_SignalHandling(t *testing.T) {
	// Test signal handling setup
	manager := NewMockConnectionManager()
	queueDef := NewQueue("test-queue")
	dispatcher := NewDispatcher(manager, []*QueueDefinition{queueDef})

	// We can't test the blocking behavior directly, but we can verify the interface
	// by ensuring ConsumeBlocking is available and doesn't panic when called
	// Note: This would block in real usage, so we don't actually call it in tests
	if dispatcher == nil {
		t.Error("Expected dispatcher to be created")
	}
}

func TestDispatcher_ComplexIntegration(t *testing.T) {
	// Test multiple registrations and complex scenarios
	manager := NewMockConnectionManager()
	channel := NewMockAMQPChannel()
	manager.SetChannel(channel)

	queueDefs := []*QueueDefinition{
		NewQueue("queue1"),
		NewQueue("queue2"),
		NewQueue("queue3").WithDQL(),
	}

	dispatcher := NewDispatcher(manager, queueDefs)

	// Register multiple consumers
	handler1 := func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil }
	handler2 := func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil }

	// Test multiple registrations
	err1 := dispatcher.RegisterByType("queue1", DispatcherTestMessage{}, handler1)
	err2 := dispatcher.RegisterByExchange("queue2", DispatcherTestMessage{}, "exchange1", handler2)
	err3 := dispatcher.RegisterByRoutingKey("queue3", DispatcherTestMessage{}, "key1", handler1)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Errorf("Expected all registrations to succeed, got errors: %v, %v, %v", err1, err2, err3)
	}

	// Test duplicate registration should fail
	err4 := dispatcher.RegisterByType("queue1", DispatcherTestMessage{}, handler1)
	if err4 == nil {
		t.Error("Expected duplicate registration to fail")
	}
	if err4.Error() != "consumer already registered for the message" {
		t.Errorf("Expected specific duplicate error, got: %v", err4)
	}
}

func TestDispatcher_ErrorHandling(t *testing.T) {
	// Test various error conditions
	manager := NewMockConnectionManager()
	queueDef := NewQueue("test-queue")
	dispatcher := NewDispatcher(manager, []*QueueDefinition{queueDef})

	handler := func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil }

	// Test registration with nonexistent queue
	err := dispatcher.RegisterByType("nonexistent-queue", DispatcherTestMessage{}, handler)
	if err == nil {
		t.Error("Expected error for nonexistent queue")
	}

	// Test registration with empty parameters
	err = dispatcher.RegisterByType("", DispatcherTestMessage{}, handler)
	if err == nil {
		t.Error("Expected error for empty queue name")
	}

	err = dispatcher.RegisterByExchange("test-queue", DispatcherTestMessage{}, "", handler)
	if err == nil {
		t.Error("Expected error for empty exchange name")
	}

	err = dispatcher.RegisterByRoutingKey("test-queue", DispatcherTestMessage{}, "", handler)
	if err == nil {
		t.Error("Expected error for empty routing key")
	}

	err = dispatcher.RegisterByExchangeRoutingKey("test-queue", DispatcherTestMessage{}, "", "key", handler)
	if err == nil {
		t.Error("Expected error for empty exchange in exchange+routing key registration")
	}

	err = dispatcher.RegisterByExchangeRoutingKey("test-queue", DispatcherTestMessage{}, "exchange", "", handler)
	if err == nil {
		t.Error("Expected error for empty routing key in exchange+routing key registration")
	}
}

func TestDispatcher_WithPointerTypes(t *testing.T) {
	// Test registration with pointer types
	manager := NewMockConnectionManager()
	queueDef := NewQueue("test-queue")
	dispatcher := NewDispatcher(manager, []*QueueDefinition{queueDef})

	handler := func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil }

	// Test pointer type registration
	err := dispatcher.RegisterByType("test-queue", &DispatcherTestMessage{}, handler)
	if err != nil {
		t.Errorf("Expected pointer type registration to succeed, got: %v", err)
	}
}

func TestDispatcher_MetadataExtractionAdvanced(t *testing.T) {
	// Test more complex metadata extraction scenarios
	manager := NewMockConnectionManager()
	dispatcher := NewDispatcher(manager, []*QueueDefinition{})

	type dispatcherImpl interface {
		extractMetadata(delivery *amqp.Delivery) (*DeliveryMetadata, error)
	}

	disp, ok := dispatcher.(dispatcherImpl)
	if !ok {
		t.Fatal("dispatcher does not implement expected interface")
	}

	// Test delivery with x-death header (retry scenario)
	delivery := &amqp.Delivery{
		MessageId:  "test-message-2",
		Type:       "DispatcherTestMessage",
		Exchange:   "dlq-exchange",
		RoutingKey: "dlq-key",
		Headers: map[string]interface{}{
			"x-death": []interface{}{
				amqp.Table{
					"count":        int64(3),
					"exchange":     "original-exchange",
					"routing-keys": []interface{}{"original.key"},
				},
			},
		},
	}

	metadata, err := disp.extractMetadata(delivery)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if metadata.XCount != 3 {
		t.Errorf("Expected XCount 3, got %d", metadata.XCount)
	}
	if metadata.OriginExchange != "original-exchange" {
		t.Errorf("Expected OriginExchange original-exchange, got %s", metadata.OriginExchange)
	}
	if metadata.RoutingKey != "original.key" {
		t.Errorf("Expected RoutingKey original.key, got %s", metadata.RoutingKey)
	}
}

// // Mock acknowledger for testing
// type mockAcknowledger struct {
// 	ackFunc  func(multiple bool) error
// 	nackFunc func(multiple, requeue bool) error
// }

// func (m *mockAcknowledger) Ack(tag uint64, multiple bool) error {
// 	if m.ackFunc != nil {
// 		return m.ackFunc(multiple)
// 	}
// 	return nil
// }

// func (m *mockAcknowledger) Nack(tag uint64, multiple, requeue bool) error {
// 	if m.nackFunc != nil {
// 		return m.nackFunc(multiple, requeue)
// 	}
// 	return nil
// }

// func (m *mockAcknowledger) Reject(tag uint64, requeue bool) error {
// 	if m.nackFunc != nil {
// 		return m.nackFunc(false, requeue)
// 	}
// 	return nil
// }
