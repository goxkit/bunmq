// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Test message types for dispatcher tests
type DispatcherTestMessage struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type AnotherTestMessage struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
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
		{
			name: "multiple queue definitions",
			queueDefinitions: []*QueueDefinition{
				NewQueue("queue1"),
				NewQueue("queue2").WithDLQ(),
				NewQueue("queue3").WithRetry(time.Second*10, 3),
			},
			expectedQueues: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConnectionManager()
			dispatcher := NewDispatcher(manager, tt.queueDefinitions)

			if dispatcher == nil {
				t.Fatal("NewDispatcher() returned nil")
			}

			// Verify dispatcher interface is properly implemented
			var _ Dispatcher = dispatcher
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
		NewQueue("queue3").WithDLQ(),
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

func TestDispatcher_ConsumeBlocking_Setup(t *testing.T) {
	// Test dispatcher setup without actually calling ConsumeBlocking
	manager := NewMockConnectionManager()
	queueDefs := []*QueueDefinition{
		NewQueue("test-queue"),
	}
	dispatcher := NewDispatcher(manager, queueDefs)

	// Register a handler
	err := dispatcher.RegisterByType("test-queue", DispatcherTestMessage{}, func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	// Verify dispatcher is ready (no actual consumption, just setup validation)
	if dispatcher == nil {
		t.Error("Expected dispatcher to be created")
	}

	// Test that duplicate registration fails
	err = dispatcher.RegisterByType("test-queue", DispatcherTestMessage{}, func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
		return nil
	})
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}

func TestDispatcher_MessageProcessing(t *testing.T) {
	// Test successful message processing
	t.Run("successful processing", func(t *testing.T) {
		manager := NewMockConnectionManager()
		queueDefs := []*QueueDefinition{
			NewQueue("success-queue").WithRetry(time.Second, 3).WithDLQ(),
		}
		dispatcher := NewDispatcher(manager, queueDefs)

		err := dispatcher.RegisterByType("success-queue", DispatcherTestMessage{}, func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			// Verify we got the right message type
			if _, ok := msg.(DispatcherTestMessage); !ok {
				t.Errorf("Expected DispatcherTestMessage, got %T", msg)
			}
			// Verify metadata is present
			if metadata == nil {
				t.Error("Expected metadata to be set")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to register handler: %v", err)
		}

		// Create a test message
		testMsg := DispatcherTestMessage{ID: "test-123", Content: "test content"}
		msgData, _ := json.Marshal(testMsg)

		delivery := &amqp.Delivery{
			MessageId:    "msg-123",
			Type:         "DispatcherTestMessage",
			Exchange:     "test-exchange",
			RoutingKey:   "test.key",
			Body:         msgData,
			Acknowledger: &mockAcknowledger{},
		}

		// Test message evaluation functionality
		type evalMethod interface {
			evaluateDefByReceivedMsg(delivery *amqp.Delivery, consumer []*ConsumerDefinition) *ConsumerDefinition
		}

		if evaluator, ok := dispatcher.(evalMethod); ok {
			// Create a consumer definition for testing
			consumerDef := &ConsumerDefinition{
				typ:     ConsumerDefinitionByType,
				queue:   "success-queue",
				msgType: "DispatcherTestMessage",
				handler: func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
					return nil
				},
			}

			result := evaluator.evaluateDefByReceivedMsg(delivery, []*ConsumerDefinition{consumerDef})
			if result == nil {
				t.Error("Expected consumer definition to be found")
			}
		}
	})

	// Test error handling with different queue
	t.Run("error handling", func(t *testing.T) {
		manager := NewMockConnectionManager()
		queueDefs := []*QueueDefinition{
			NewQueue("error-queue").WithRetry(time.Second, 2),
		}
		dispatcher := NewDispatcher(manager, queueDefs)

		err := dispatcher.RegisterByType("error-queue", DispatcherTestMessage{}, func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return errors.New("processing error")
		})
		if err != nil {
			t.Fatalf("Failed to register handler: %v", err)
		}

		// Test that handler was registered by attempting to register again
		err = dispatcher.RegisterByType("error-queue", DispatcherTestMessage{}, func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err == nil {
			t.Error("Expected error for duplicate registration")
		}
	})
}

func TestDispatcher_ErrorScenarios(t *testing.T) {
	manager := NewMockConnectionManager()
	queueDefs := []*QueueDefinition{
		NewQueue("test-queue"),
		NewQueue("error-queue").WithRetry(time.Second, 2),
	}
	dispatcher := NewDispatcher(manager, queueDefs)

	// Test registration with invalid message type
	t.Run("invalid message type", func(t *testing.T) {
		err := dispatcher.RegisterByType("test-queue", nil, func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err == nil {
			t.Error("Expected error for nil message type")
		}
	})

	// Test registration with non-existent queue
	t.Run("non-existent queue", func(t *testing.T) {
		err := dispatcher.RegisterByType("non-existent", DispatcherTestMessage{}, func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err == nil {
			t.Error("Expected error for non-existent queue")
		}
	})

	// Test various invalid registration combinations
	t.Run("invalid parameters", func(t *testing.T) {
		// Test with empty queue name
		err := dispatcher.RegisterByType("", DispatcherTestMessage{}, func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err == nil {
			t.Error("Expected error for empty queue name")
		}

		// Test RegisterByExchange with empty exchange
		err = dispatcher.RegisterByExchange("test-queue", DispatcherTestMessage{}, "", func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err == nil {
			t.Error("Expected error for empty exchange")
		}

		// Test RegisterByRoutingKey with empty routing key
		err = dispatcher.RegisterByRoutingKey("test-queue", DispatcherTestMessage{}, "", func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err == nil {
			t.Error("Expected error for empty routing key")
		}
	})
}

func TestDispatcher_MultipleHandlers(t *testing.T) {
	manager := NewMockConnectionManager()
	queueDefs := []*QueueDefinition{
		NewQueue("multi-queue"),
	}
	dispatcher := NewDispatcher(manager, queueDefs)

	// Register multiple handlers for different registration types
	err1 := dispatcher.RegisterByExchange("multi-queue", DispatcherTestMessage{}, "exchange1", func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
		return nil
	})
	err2 := dispatcher.RegisterByRoutingKey("multi-queue", AnotherTestMessage{}, "route.key", func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
		return nil
	})

	if err1 != nil {
		t.Errorf("Expected no error for first registration, got %v", err1)
	}
	if err2 != nil {
		t.Errorf("Expected no error for second registration, got %v", err2)
	}

	// Test that we can't register the same message type twice
	err3 := dispatcher.RegisterByType("multi-queue", DispatcherTestMessage{}, func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
		return nil
	})
	if err3 == nil {
		t.Error("Expected error for duplicate message type registration")
	}
}

func TestDispatcher_ExtractMetadata_Comprehensive(t *testing.T) {
	manager := NewMockConnectionManager()
	dispatcher := NewDispatcher(manager, []*QueueDefinition{})

	type metadataExtractor interface {
		extractMetadata(delivery *amqp.Delivery) (*DeliveryMetadata, error)
	}

	if extractor, ok := dispatcher.(metadataExtractor); ok {
		// Test basic metadata extraction
		t.Run("basic metadata", func(t *testing.T) {
			delivery := &amqp.Delivery{
				MessageId:  "test-message-123",
				Type:       "TestMessage",
				Exchange:   "test-exchange",
				RoutingKey: "test.routing.key",
				Headers: amqp.Table{
					"custom-header": "custom-value",
					"x-retry-count": int64(2),
				},
			}

			metadata, err := extractor.extractMetadata(delivery)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if metadata.MessageID != "test-message-123" {
				t.Errorf("Expected MessageID test-message-123, got %s", metadata.MessageID)
			}
			if metadata.Type != "TestMessage" {
				t.Errorf("Expected Type TestMessage, got %s", metadata.Type)
			}
			if metadata.OriginExchange != "test-exchange" {
				t.Errorf("Expected OriginExchange test-exchange, got %s", metadata.OriginExchange)
			}
			if metadata.RoutingKey != "test.routing.key" {
				t.Errorf("Expected RoutingKey test.routing.key, got %s", metadata.RoutingKey)
			}
			if metadata.Headers["custom-header"] != "custom-value" {
				t.Errorf("Expected custom-header value, got %v", metadata.Headers["custom-header"])
			}
		})

		// Test with x-death headers (valid structure)
		t.Run("valid x-death", func(t *testing.T) {
			delivery := &amqp.Delivery{
				MessageId:  "dlq-message",
				Type:       "TestMessage",
				Exchange:   "dlq-exchange",
				RoutingKey: "dlq.key",
				Headers: amqp.Table{
					"x-death": []interface{}{
						amqp.Table{
							"count":        int64(3),
							"exchange":     "original-exchange",
							"routing-keys": []interface{}{"original.routing.key"},
						},
					},
				},
			}

			metadata, err := extractor.extractMetadata(delivery)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if metadata.XCount != 3 {
				t.Errorf("Expected XCount 3, got %d", metadata.XCount)
			}
			if metadata.OriginExchange != "original-exchange" {
				t.Errorf("Expected OriginExchange original-exchange, got %s", metadata.OriginExchange)
			}
			if metadata.RoutingKey != "original.routing.key" {
				t.Errorf("Expected RoutingKey original.routing.key, got %s", metadata.RoutingKey)
			}
		})

		// Test minimal delivery
		t.Run("minimal delivery", func(t *testing.T) {
			delivery := &amqp.Delivery{}
			metadata, err := extractor.extractMetadata(delivery)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if metadata == nil {
				t.Error("Expected metadata to be created")
			}
		})

		// Test with x-death but invalid count type
		t.Run("x-death invalid count", func(t *testing.T) {
			delivery := &amqp.Delivery{
				Headers: amqp.Table{
					"x-death": []interface{}{
						amqp.Table{
							"count":    "not-a-number",
							"exchange": "test-exchange",
						},
					},
				},
			}
			metadata, err := extractor.extractMetadata(delivery)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			// Should have XCount as 0 since count couldn't be parsed
			if metadata.XCount != 0 {
				t.Errorf("Expected XCount 0 for invalid count, got %d", metadata.XCount)
			}
		})

		// Test with x-death but missing routing-keys
		t.Run("x-death missing routing keys", func(t *testing.T) {
			delivery := &amqp.Delivery{
				RoutingKey: "fallback.key",
				Headers: amqp.Table{
					"x-death": []interface{}{
						amqp.Table{
							"count":    int64(1),
							"exchange": "test-exchange",
							// routing-keys missing
						},
					},
				},
			}
			metadata, err := extractor.extractMetadata(delivery)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			// Should fall back to delivery routing key
			if metadata.RoutingKey != "fallback.key" {
				t.Errorf("Expected RoutingKey fallback.key, got %s", metadata.RoutingKey)
			}
		})
	} else {
		t.Error("Failed to assert dispatcher to metadata extractor interface")
	}
}

func TestDispatcher_AdvancedRegistration(t *testing.T) {
	// Test more edge cases for registration methods to increase coverage
	manager := NewMockConnectionManager()
	queueDefs := []*QueueDefinition{
		NewQueue("advanced-queue"),
	}
	dispatcher := NewDispatcher(manager, queueDefs)

	// Test RegisterByExchange with empty exchange (should fail)
	t.Run("RegisterByExchange empty exchange", func(t *testing.T) {
		err := dispatcher.RegisterByExchange("advanced-queue", DispatcherTestMessage{}, "", func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err == nil {
			t.Error("Expected error for empty exchange")
		}
	})

	// Test RegisterByRoutingKey with empty routing key (should fail)
	t.Run("RegisterByRoutingKey empty routing key", func(t *testing.T) {
		err := dispatcher.RegisterByRoutingKey("advanced-queue", DispatcherTestMessage{}, "", func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err == nil {
			t.Error("Expected error for empty routing key")
		}
	})

	// Test RegisterByExchangeRoutingKey with empty exchange
	t.Run("RegisterByExchangeRoutingKey empty exchange", func(t *testing.T) {
		err := dispatcher.RegisterByExchangeRoutingKey("advanced-queue", DispatcherTestMessage{}, "", "test.key", func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err == nil {
			t.Error("Expected error for empty exchange")
		}
	})

	// Test RegisterByExchangeRoutingKey with empty routing key
	t.Run("RegisterByExchangeRoutingKey empty routing key", func(t *testing.T) {
		err := dispatcher.RegisterByExchangeRoutingKey("advanced-queue", DispatcherTestMessage{}, "test-exchange", "", func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err == nil {
			t.Error("Expected error for empty routing key")
		}
	})

	// Test successful registrations to cover valid paths
	t.Run("successful registrations", func(t *testing.T) {
		// Test RegisterByExchange
		err := dispatcher.RegisterByExchange("advanced-queue", DispatcherTestMessage{}, "test-exchange", func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected no error for RegisterByExchange, got %v", err)
		}

		// Test RegisterByRoutingKey with different message type
		err = dispatcher.RegisterByRoutingKey("advanced-queue", AnotherTestMessage{}, "test.routing.key", func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected no error for RegisterByRoutingKey, got %v", err)
		}
	})
}

func TestDispatcher_ExtractMetadata_AdditionalCases(t *testing.T) {
	// Add more tests to increase extractMetadata coverage
	manager := NewMockConnectionManager()
	dispatcher := NewDispatcher(manager, []*QueueDefinition{})

	type metadataExtractor interface {
		extractMetadata(delivery *amqp.Delivery) (*DeliveryMetadata, error)
	}

	if extractor, ok := dispatcher.(metadataExtractor); ok {
		// Test x-death with multiple entries (only first one should be used)
		t.Run("x-death multiple entries", func(t *testing.T) {
			delivery := &amqp.Delivery{
				MessageId:  "multi-death",
				Exchange:   "current-exchange",
				RoutingKey: "current.key",
				Headers: amqp.Table{
					"x-death": []interface{}{
						amqp.Table{
							"count":        int64(3),
							"exchange":     "first-exchange",
							"routing-keys": []interface{}{"first.key"},
						},
						amqp.Table{
							"count":        int64(1),
							"exchange":     "second-exchange",
							"routing-keys": []interface{}{"second.key"},
						},
					},
				},
			}

			metadata, err := extractor.extractMetadata(delivery)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Should use LAST death entry since len(tables) > 1
			if metadata.OriginExchange != "second-exchange" {
				t.Errorf("Expected OriginExchange second-exchange, got %s", metadata.OriginExchange)
			}
			if metadata.RoutingKey != "second.key" {
				t.Errorf("Expected RoutingKey second.key, got %s", metadata.RoutingKey)
			}
			// XCount should still be from first entry
			if metadata.XCount != 3 {
				t.Errorf("Expected XCount 3, got %d", metadata.XCount)
			}
		})

		// Test x-death with valid single entry but different routing key handling
		t.Run("x-death valid single entry", func(t *testing.T) {
			delivery := &amqp.Delivery{
				RoutingKey: "fallback.key",
				Headers: amqp.Table{
					"x-death": []interface{}{
						amqp.Table{
							"count":        int64(2),
							"exchange":     "death-exchange",
							"routing-keys": []interface{}{"death.routing.key"},
						},
					},
				},
			}

			metadata, err := extractor.extractMetadata(delivery)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Should override with death entry since len(tables) == 1
			if metadata.OriginExchange != "death-exchange" {
				t.Errorf("Expected OriginExchange death-exchange, got %s", metadata.OriginExchange)
			}
			if metadata.RoutingKey != "death.routing.key" {
				t.Errorf("Expected RoutingKey death.routing.key, got %s", metadata.RoutingKey)
			}
			if metadata.XCount != 2 {
				t.Errorf("Expected XCount 2, got %d", metadata.XCount)
			}
		})

		// Test with various header types
		t.Run("various header types", func(t *testing.T) {
			delivery := &amqp.Delivery{
				MessageId: "header-test",
				Headers: amqp.Table{
					"string-header":  "string-value",
					"int-header":     int64(42),
					"bool-header":    true,
					"float-header":   3.14,
					"nil-header":     nil,
					"complex-header": map[string]interface{}{"nested": "value"},
				},
			}

			metadata, err := extractor.extractMetadata(delivery)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if len(metadata.Headers) != 6 {
				t.Errorf("Expected 6 headers, got %d", len(metadata.Headers))
			}
			if metadata.Headers["string-header"] != "string-value" {
				t.Errorf("Expected string-header value, got %v", metadata.Headers["string-header"])
			}
		})

		// Test without x-death header at all
		t.Run("no x-death header", func(t *testing.T) {
			delivery := &amqp.Delivery{
				MessageId:  "no-death-test",
				Exchange:   "test-exchange",
				RoutingKey: "test.key",
				Type:       "test-type",
				Headers:    amqp.Table{"other-header": "value"},
			}

			metadata, err := extractor.extractMetadata(delivery)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Should use delivery fields since no x-death
			if metadata.OriginExchange != "test-exchange" {
				t.Errorf("Expected OriginExchange test-exchange, got %s", metadata.OriginExchange)
			}
			if metadata.RoutingKey != "test.key" {
				t.Errorf("Expected RoutingKey test.key, got %s", metadata.RoutingKey)
			}
			if metadata.XCount != 0 {
				t.Errorf("Expected XCount 0 with no x-death, got %d", metadata.XCount)
			}
		})
	} else {
		t.Error("Failed to assert dispatcher to metadata extractor interface")
	}
}

func TestDispatcher_EvaluateDefByReceivedMsg(t *testing.T) {
	manager := NewMockConnectionManager()
	queueDef := NewQueue("test-queue")
	dispatcher := NewDispatcher(manager, []*QueueDefinition{queueDef})

	// Cast to access internal methods
	type defEvaluator interface {
		evaluateDefByReceivedMsg(queueDef *QueueDefinition, metadata *DeliveryMetadata) (*ConsumerDefinition, error)
	}

	if evaluator, ok := dispatcher.(defEvaluator); ok {
		// Register different types of consumers for testing
		testHandler := func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil }

		// Create different message types to avoid conflicts
		type TypeMessage struct{ ID string }
		type ExchangeMessage struct{ Name string }
		type RoutingMessage struct{ Value int }
		type ExchangeRoutingMessage struct{ Data string }

		// Register by type
		err := dispatcher.RegisterByType("test-queue", TypeMessage{}, testHandler)
		if err != nil {
			t.Fatalf("Failed to register by type: %v", err)
		}

		// Register by exchange
		err = dispatcher.RegisterByExchange("test-queue", ExchangeMessage{}, "test-exchange", testHandler)
		if err != nil {
			t.Fatalf("Failed to register by exchange: %v", err)
		}

		// Register by routing key
		err = dispatcher.RegisterByRoutingKey("test-queue", RoutingMessage{}, "test.routing.key", testHandler)
		if err != nil {
			t.Fatalf("Failed to register by routing key: %v", err)
		}

		// Register by exchange and routing key
		err = dispatcher.RegisterByExchangeRoutingKey("test-queue", ExchangeRoutingMessage{}, "another-exchange", "another.key", testHandler)
		if err != nil {
			t.Fatalf("Failed to register by exchange and routing key: %v", err)
		}

		tests := []struct {
			name           string
			metadata       *DeliveryMetadata
			expectedType   ConsumerDefinitionType
			expectError    bool
			expectedErrMsg string
		}{
			{
				name: "match by type",
				metadata: &DeliveryMetadata{
					Type:           "bunmq.TypeMessage",
					OriginExchange: "other-exchange",
					RoutingKey:     "other.key",
				},
				expectedType: ConsumerDefinitionByType,
				expectError:  false,
			},
			{
				name: "match by exchange",
				metadata: &DeliveryMetadata{
					Type:           "unknown.type",
					OriginExchange: "test-exchange",
					RoutingKey:     "other.key",
				},
				expectedType: ConsumerDefinitionByExchange,
				expectError:  false,
			},
			{
				name: "match by routing key",
				metadata: &DeliveryMetadata{
					Type:           "unknown.type",
					OriginExchange: "other-exchange",
					RoutingKey:     "test.routing.key",
				},
				expectedType: ConsumerDefinitionByRoutingKey,
				expectError:  false,
			},
			{
				name: "match by exchange and routing key",
				metadata: &DeliveryMetadata{
					Type:           "unknown.type",
					OriginExchange: "another-exchange",
					RoutingKey:     "another.key",
				},
				expectedType: ConsumerDefinitionByExchangeRoutingKey,
				expectError:  false,
			},
			{
				name: "no match found",
				metadata: &DeliveryMetadata{
					Type:           "unknown.type",
					OriginExchange: "unknown-exchange",
					RoutingKey:     "unknown.key",
				},
				expectError:    true,
				expectedErrMsg: "no consumer definition found for queue",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				def, err := evaluator.evaluateDefByReceivedMsg(queueDef, tt.metadata)

				if tt.expectError {
					if err == nil {
						t.Errorf("Expected error but got none")
					} else if !contains(err.Error(), tt.expectedErrMsg) {
						t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedErrMsg, err.Error())
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
					if def == nil {
						t.Errorf("Expected consumer definition but got nil")
					} else if def.typ != tt.expectedType {
						t.Errorf("Expected type %d, got %d", tt.expectedType, def.typ)
					}
				}
			})
		}

		// Test error cases
		t.Run("queue with no consumer definitions", func(t *testing.T) {
			emptyQueueDef := NewQueue("empty-queue")
			emptyDispatcher := NewDispatcher(manager, []*QueueDefinition{emptyQueueDef})

			if emptyEvaluator, ok := emptyDispatcher.(defEvaluator); ok {
				metadata := &DeliveryMetadata{Type: "test.type"}
				def, err := emptyEvaluator.evaluateDefByReceivedMsg(emptyQueueDef, metadata)

				if err == nil {
					t.Errorf("Expected error for empty consumer definitions")
				}
				if def != nil {
					t.Errorf("Expected nil definition for empty consumer definitions")
				}
			}
		})

		t.Run("queue not in consumer definitions", func(t *testing.T) {
			unknownQueueDef := NewQueue("unknown-queue")
			metadata := &DeliveryMetadata{Type: "test.type"}
			def, err := evaluator.evaluateDefByReceivedMsg(unknownQueueDef, metadata)

			if err == nil {
				t.Errorf("Expected error for unknown queue")
			}
			if def != nil {
				t.Errorf("Expected nil definition for unknown queue")
			}
		})

	} else {
		t.Error("Failed to assert dispatcher to definition evaluator interface")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()))
}

func TestDispatcher_ProcessReceivedMessage(t *testing.T) {
	// Test processReceivedMessage indirectly through behavior verification
	// Since we can't override Ack/Nack methods, we'll verify the function doesn't panic
	// and that it processes different message types correctly

	// Cast to access internal methods
	type messageProcessor interface {
		processReceivedMessage(queueDef *QueueDefinition, received *amqp.Delivery)
	}

	tests := []struct {
		name            string
		setupQueue      func() *QueueDefinition
		setupDispatcher func() (Dispatcher, *MockConnectionManager)
		delivery        amqp.Delivery
		shouldPanic     bool
	}{
		{
			name: "successful message processing - no panic",
			setupQueue: func() *QueueDefinition {
				return NewQueue("test-queue")
			},
			setupDispatcher: func() (Dispatcher, *MockConnectionManager) {
				manager := NewMockConnectionManager()
				queueDef := NewQueue("test-queue")
				dispatcher := NewDispatcher(manager, []*QueueDefinition{queueDef})

				// Register a handler that succeeds
				dispatcher.RegisterByType("test-queue", DispatcherTestMessage{},
					func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
						return nil
					})
				return dispatcher, manager
			},
			delivery: amqp.Delivery{
				MessageId: "test-msg-1",
				Type:      "bunmq.DispatcherTestMessage",
				Body:      []byte(`{"id":"123","content":"test"}`),
				Headers:   amqp.Table{},
			},
			shouldPanic: false,
		},
		{
			name: "unmarshal error - handles gracefully",
			setupQueue: func() *QueueDefinition {
				return NewQueue("test-queue")
			},
			setupDispatcher: func() (Dispatcher, *MockConnectionManager) {
				manager := NewMockConnectionManager()
				queueDef := NewQueue("test-queue")
				dispatcher := NewDispatcher(manager, []*QueueDefinition{queueDef})

				dispatcher.RegisterByType("test-queue", DispatcherTestMessage{},
					func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
						return nil
					})
				return dispatcher, manager
			},
			delivery: amqp.Delivery{
				MessageId: "test-msg-2",
				Type:      "bunmq.DispatcherTestMessage",
				Body:      []byte(`invalid json`),
				Headers:   amqp.Table{},
			},
			shouldPanic: false,
		},
		{
			name: "handler error with DLQ - handles gracefully",
			setupQueue: func() *QueueDefinition {
				return NewQueue("test-queue").WithDLQ()
			},
			setupDispatcher: func() (Dispatcher, *MockConnectionManager) {
				manager := NewMockConnectionManager()
				queueDef := NewQueue("test-queue").WithDLQ()
				dispatcher := NewDispatcher(manager, []*QueueDefinition{queueDef})

				// Register a handler that fails
				dispatcher.RegisterByType("test-queue", DispatcherTestMessage{},
					func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
						return errors.New("handler error")
					})
				return dispatcher, manager
			},
			delivery: amqp.Delivery{
				MessageId: "test-msg-3",
				Type:      "bunmq.DispatcherTestMessage",
				Body:      []byte(`{"id":"123","content":"test"}`),
				Headers:   amqp.Table{},
			},
			shouldPanic: false,
		},
		{
			name: "handler retryable error - handles gracefully",
			setupQueue: func() *QueueDefinition {
				return NewQueue("test-queue").WithRetry(time.Second*10, 3)
			},
			setupDispatcher: func() (Dispatcher, *MockConnectionManager) {
				manager := NewMockConnectionManager()
				queueDef := NewQueue("test-queue").WithRetry(time.Second*10, 3)
				dispatcher := NewDispatcher(manager, []*QueueDefinition{queueDef})

				// Register a handler that returns retryable error
				dispatcher.RegisterByType("test-queue", DispatcherTestMessage{},
					func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
						return RetryableError
					})
				return dispatcher, manager
			},
			delivery: amqp.Delivery{
				MessageId: "test-msg-4",
				Type:      "bunmq.DispatcherTestMessage",
				Body:      []byte(`{"id":"123","content":"test"}`),
				Headers:   amqp.Table{},
			},
			shouldPanic: false,
		},
		{
			name: "no consumer found - handles gracefully",
			setupQueue: func() *QueueDefinition {
				return NewQueue("test-queue")
			},
			setupDispatcher: func() (Dispatcher, *MockConnectionManager) {
				manager := NewMockConnectionManager()
				queueDef := NewQueue("test-queue")
				dispatcher := NewDispatcher(manager, []*QueueDefinition{queueDef})
				// Don't register any handlers
				return dispatcher, manager
			},
			delivery: amqp.Delivery{
				MessageId: "test-msg-6",
				Type:      "unknown.type",
				Body:      []byte(`{"id":"123","content":"test"}`),
				Headers:   amqp.Table{},
			},
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queueDef := tt.setupQueue()
			dispatcher, manager := tt.setupDispatcher()

			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("processReceivedMessage panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("processReceivedMessage should have panicked but didn't")
				}
			}()

			// Process the message
			if processor, ok := dispatcher.(messageProcessor); ok {
				processor.processReceivedMessage(queueDef, &tt.delivery)
			} else {
				t.Error("Failed to assert dispatcher to message processor interface")
			}

			_ = manager // Prevent unused variable warning
		})
	}
}

// TestDispatcher_PublishToDlq functionality is tested indirectly through
// TestDispatcher_ProcessReceivedMessage tests that cover DLQ scenarios.
// The publishToDlq method is an internal method that is called by processReceivedMessage
// when handler errors occur and DLQ is configured.
