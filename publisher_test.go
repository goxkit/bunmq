// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// MockConnectionManager is a mock implementation of ConnectionManager interface for testing
type MockConnectionManager struct {
	connection        RMQConnection
	channel           AMQPChannel
	connectionString  string
	closed            bool
	healthy           bool
	getConnectionErr  error
	getChannelErr     error
	closeErr          error
	reconnectCallback func(RMQConnection, AMQPChannel)
	topology          *topology
}

func NewMockConnectionManager() *MockConnectionManager {
	return &MockConnectionManager{
		connection:       NewMockRMQConnection(),
		channel:          NewMockAMQPChannel(),
		connectionString: "amqp://test",
		closed:           false,
		healthy:          true,
	}
}

func (m *MockConnectionManager) GetConnection() (RMQConnection, error) {
	if m.getConnectionErr != nil {
		return nil, m.getConnectionErr
	}
	return m.connection, nil
}

func (m *MockConnectionManager) GetChannel() (AMQPChannel, error) {
	if m.getChannelErr != nil {
		return nil, m.getChannelErr
	}
	return m.channel, nil
}

func (m *MockConnectionManager) GetConnectionString() string {
	return m.connectionString
}

func (m *MockConnectionManager) Close() error {
	m.closed = true
	return m.closeErr
}

func (m *MockConnectionManager) IsHealthy() bool {
	return m.healthy
}

func (m *MockConnectionManager) SetReconnectCallback(callback func(conn RMQConnection, ch AMQPChannel)) {
	m.reconnectCallback = callback
}

func (m *MockConnectionManager) SetTopology(t Topology) {
	m.topology = t.(*topology)
}

// Helper methods for testing
func (m *MockConnectionManager) SetConnection(conn RMQConnection) {
	m.connection = conn
}

func (m *MockConnectionManager) SetChannel(ch AMQPChannel) {
	m.channel = ch
}

func (m *MockConnectionManager) SetConnectionString(cs string) {
	m.connectionString = cs
}

func (m *MockConnectionManager) SetHealthy(healthy bool) {
	m.healthy = healthy
}

func (m *MockConnectionManager) SetGetConnectionError(err error) {
	m.getConnectionErr = err
}

func (m *MockConnectionManager) SetGetChannelError(err error) {
	m.getChannelErr = err
}

func (m *MockConnectionManager) SetCloseError(err error) {
	m.closeErr = err
}

func (m *MockConnectionManager) Qos(prefetchCount, prefetchSize int, global bool) error {
	return nil
}

// Test message types
type TestMessage struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type TestPointerMessage struct {
	Value int `json:"value"`
}

func TestNewPublisher(t *testing.T) {
	manager := NewMockConnectionManager()
	appName := "test-app"

	publisher := NewPublisher(appName, manager)
	if publisher == nil {
		t.Fatal("NewPublisher() returned nil")
	}

	// NewPublisher returns Publisher interface, so type check is implicit
}

func TestPublisher_Publish_SimpleCase(t *testing.T) {
	// Test publishing to default exchange (simulating SimplePublish behavior)
	manager := NewMockConnectionManager()
	channel := NewMockAMQPChannel()
	manager.SetChannel(channel)

	publisher := NewPublisher("test-app", manager)
	// Use empty exchange to simulate publishing to default exchange (direct to queue)
	err := publisher.Publish(context.Background(), "", "test-queue", TestMessage{ID: "123", Content: "test"})

	// This should fail because of empty exchange validation
	if err == nil {
		t.Error("Publish with empty exchange should return error")
	}
}

func TestPublisher_Publish(t *testing.T) {
	tests := []struct {
		name        string
		exchange    string
		routingKey  string
		msg         interface{}
		channelErr  error
		publishErr  error
		expectError bool
	}{
		{
			name:        "successful publish",
			exchange:    "test-exchange",
			routingKey:  "test.key",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  nil,
			publishErr:  nil,
			expectError: false,
		},
		{
			name:        "empty exchange",
			exchange:    "",
			routingKey:  "test.key",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  nil,
			publishErr:  nil,
			expectError: true,
		},
		{
			name:        "channel error",
			exchange:    "test-exchange",
			routingKey:  "test.key",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  errors.New("channel error"),
			publishErr:  nil,
			expectError: true,
		},
		{
			name:        "publish error",
			exchange:    "test-exchange",
			routingKey:  "test.key",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  nil,
			publishErr:  errors.New("publish error"),
			expectError: true,
		},
		{
			name:        "empty routing key",
			exchange:    "test-exchange",
			routingKey:  "",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  nil,
			publishErr:  nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConnectionManager()
			channel := NewMockAMQPChannel()

			if tt.channelErr != nil {
				manager.SetGetChannelError(tt.channelErr)
			} else {
				if tt.publishErr != nil {
					channel.SetPublishError(tt.publishErr)
				}
				manager.SetChannel(channel)
			}

			publisher := NewPublisher("test-app", manager)
			err := publisher.Publish(context.Background(), tt.exchange, tt.routingKey, tt.msg)

			if tt.expectError {
				if err == nil {
					t.Error("Publish() should return error")
				}
			} else {
				if err != nil {
					t.Errorf("Publish() returned unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPublisher_PublishDeadline(t *testing.T) {
	tests := []struct {
		name        string
		exchange    string
		routingKey  string
		msg         interface{}
		channelErr  error
		publishErr  error
		expectError bool
	}{
		{
			name:        "timeout waiting for confirmation",
			exchange:    "test-exchange",
			routingKey:  "test.key",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  nil,
			publishErr:  nil,
			expectError: true, // Expect timeout error since no real broker confirmation
		},
		{
			name:        "empty exchange with deadline",
			exchange:    "",
			routingKey:  "test.key",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  nil,
			publishErr:  nil,
			expectError: true,
		},
		{
			name:        "channel error with deadline",
			exchange:    "test-exchange",
			routingKey:  "test.key",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  errors.New("channel error"),
			publishErr:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConnectionManager()
			channel := NewMockAMQPChannel()
			// Don't set DeferredConfirmation - let it return nil to simulate missing broker confirmation

			if tt.channelErr != nil {
				manager.SetGetChannelError(tt.channelErr)
			} else {
				if tt.publishErr != nil {
					channel.SetPublishError(tt.publishErr)
				}
				manager.SetChannel(channel)
			}

			publisher := NewPublisher("test-app", manager)
			err := publisher.PublishDeadline(context.Background(), tt.exchange, tt.routingKey, tt.msg)

			if tt.expectError {
				if err == nil {
					t.Error("PublishDeadline() should return error")
				}
				// For the timeout case, check that we get a reasonable error message
				if tt.name == "timeout waiting for confirmation" && err != nil {
					t.Logf("PublishDeadline() correctly returned timeout/confirmation error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("PublishDeadline() returned unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPublisher_PublishQueue(t *testing.T) {
	tests := []struct {
		name        string
		queue       string
		msg         interface{}
		channelErr  error
		publishErr  error
		expectError bool
	}{
		{
			name:        "successful queue publish",
			queue:       "test-queue",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  nil,
			publishErr:  nil,
			expectError: false,
		},
		{
			name:        "empty queue",
			queue:       "",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  nil,
			publishErr:  nil,
			expectError: true,
		},
		{
			name:        "channel error",
			queue:       "test-queue",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  errors.New("channel error"),
			publishErr:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConnectionManager()
			channel := NewMockAMQPChannel()

			if tt.channelErr != nil {
				manager.SetGetChannelError(tt.channelErr)
			} else {
				if tt.publishErr != nil {
					channel.SetPublishError(tt.publishErr)
				}
				manager.SetChannel(channel)
			}

			publisher := NewPublisher("test-app", manager)
			err := publisher.PublishQueue(context.Background(), tt.queue, tt.msg)

			if tt.expectError {
				if err == nil {
					t.Error("PublishQueue() should return error")
				}
			} else {
				if err != nil {
					t.Errorf("PublishQueue() returned unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPublisher_PublishQueueDeadline(t *testing.T) {
	tests := []struct {
		name        string
		queue       string
		msg         interface{}
		channelErr  error
		publishErr  error
		expectError bool
	}{
		{
			name:        "successful queue publish with deadline",
			queue:       "test-queue",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  nil,
			publishErr:  nil,
			expectError: false,
		},
		{
			name:        "empty queue with deadline",
			queue:       "",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  nil,
			publishErr:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConnectionManager()
			channel := NewMockAMQPChannel()

			if tt.channelErr != nil {
				manager.SetGetChannelError(tt.channelErr)
			} else {
				if tt.publishErr != nil {
					channel.SetPublishError(tt.publishErr)
				}
				manager.SetChannel(channel)
			}

			publisher := NewPublisher("test-app", manager)
			err := publisher.PublishQueueDeadline(context.Background(), tt.queue, tt.msg)

			if tt.expectError {
				if err == nil {
					t.Error("PublishQueueDeadline() should return error")
				}
			} else {
				if err != nil {
					t.Errorf("PublishQueueDeadline() returned unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPublisher_MessageSerialization(t *testing.T) {
	tests := []struct {
		name    string
		msg     interface{}
		wantErr bool
	}{
		{
			name:    "struct message",
			msg:     TestMessage{ID: "123", Content: "test"},
			wantErr: false,
		},
		{
			name:    "pointer to struct",
			msg:     &TestPointerMessage{Value: 42},
			wantErr: false,
		},
		{
			name:    "string message",
			msg:     "simple string",
			wantErr: false,
		},
		{
			name:    "int message",
			msg:     123,
			wantErr: false,
		},
		{
			name:    "map message",
			msg:     map[string]interface{}{"key": "value"},
			wantErr: false,
		},
		{
			name:    "empty struct",
			msg:     struct{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConnectionManager()
			channel := NewMockAMQPChannel()
			manager.SetChannel(channel)

			publisher := NewPublisher("test-app", manager)
			err := publisher.Publish(context.Background(), "test-exchange", "test.key", tt.msg)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error for message serialization")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for message serialization: %v", err)
				}
			}
		})
	}
}

func TestPublisher_ContextCancellation(t *testing.T) {
	manager := NewMockConnectionManager()
	channel := NewMockAMQPChannel()
	manager.SetChannel(channel)

	publisher := NewPublisher("test-app", manager)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := publisher.Publish(ctx, "test-exchange", "test.key", TestMessage{ID: "123", Content: "test"})
	// The actual behavior depends on when the context is checked,
	// but we should be able to publish successfully since context cancellation
	// is mainly used for timeouts in PublishDeadline
	if err != nil {
		t.Logf("Publish with cancelled context returned: %v", err)
	}
}

func TestPublisher_Interface(t *testing.T) {
	// Test that publisher implements Publisher interface
	manager := NewMockConnectionManager()
	pub := NewPublisher("test-app", manager)

	// Test all interface methods exist
	ctx := context.Background()
	msg := TestMessage{ID: "123", Content: "test"}

	// These calls will fail because of mock setup, but we're testing interface compliance
	_ = pub.Publish(ctx, "exchange", "key", msg)
	_ = pub.PublishDeadline(ctx, "exchange", "key", msg)
	_ = pub.PublishQueue(ctx, "queue", msg)
	_ = pub.PublishQueueDeadline(ctx, "queue", msg)
}

func TestJSONContentType(t *testing.T) {
	if JSONContentType != "application/json" {
		t.Errorf("JSONContentType = %v, want application/json", JSONContentType)
	}
}

func TestPublisher_TimeoutBehavior(t *testing.T) {
	// Test that PublishDeadline completes quickly with mock (no real broker delays)
	manager := NewMockConnectionManager()
	channel := NewMockAMQPChannel()
	manager.SetChannel(channel)

	publisher := NewPublisher("test-app", manager)

	start := time.Now()
	err := publisher.PublishDeadline(context.Background(), "test-exchange", "test.key", TestMessage{ID: "123", Content: "test"})
	duration := time.Since(start)

	// Should complete quickly since mock doesn't introduce delay
	if duration > time.Second {
		t.Errorf("PublishDeadline took too long: %v", duration)
	}

	// Expect error due to mock not having real broker confirmation
	if err == nil {
		t.Error("PublishDeadline should return error when no real confirmation available")
	} else {
		t.Logf("PublishDeadline correctly returned error: %v", err)
	}
}

func TestPublisher_PublishWithOptions(t *testing.T) {
	tests := []struct {
		name         string
		options      []*Option
		expectedDM   uint8          // expected delivery mode
		expectedHdrs map[string]any // expected headers
	}{
		{
			name:         "no options - default values",
			options:      []*Option{},
			expectedDM:   0, // default
			expectedHdrs: map[string]any{},
		},
		{
			name:         "persistent delivery mode only",
			options:      NewOption().WithDeliveryMode(amqp091.Persistent).Build(),
			expectedDM:   amqp091.Persistent,
			expectedHdrs: map[string]any{},
		},
		{
			name:         "transient delivery mode only",
			options:      NewOption().WithDeliveryMode(amqp091.Transient).Build(),
			expectedDM:   amqp091.Transient,
			expectedHdrs: map[string]any{},
		},
		{
			name:         "headers only",
			options:      NewOption().WithHeaders(map[string]any{"priority": 5, "source": "test"}).Build(),
			expectedDM:   0, // default
			expectedHdrs: map[string]any{"priority": 5, "source": "test"},
		},
		{
			name: "both delivery mode and headers",
			options: append(
				NewOption().WithDeliveryMode(amqp091.Persistent).Build(),
				NewOption().WithHeaders(map[string]any{"content-type": "application/xml", "retry-count": 3}).Build()...,
			),
			expectedDM:   amqp091.Persistent,
			expectedHdrs: map[string]any{"content-type": "application/xml", "retry-count": 3},
		},
		{
			name:         "publisher options helper",
			options:      NewOption().WithDeliveryMode(amqp091.Transient).WithHeaders(map[string]any{"correlation-id": "abc123", "message-version": 2}).Build(),
			expectedDM:   amqp091.Transient,
			expectedHdrs: map[string]any{"correlation-id": "abc123", "message-version": 2},
		},
		{
			name:         "empty headers map",
			options:      NewOption().WithDeliveryMode(amqp091.Transient).WithHeaders(map[string]any{}).Build(),
			expectedDM:   amqp091.Transient,
			expectedHdrs: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConnectionManager()
			channel := NewMockAMQPChannel()
			manager.SetChannel(channel)

			publisher := NewPublisher("test-app", manager)

			// Clear any previous messages
			channel.ClearPublishedMessages()

			err := publisher.Publish(
				context.Background(),
				"test-exchange",
				"test.key",
				TestMessage{ID: "123", Content: "test"},
				tt.options...,
			)

			if err != nil {
				t.Errorf("Publish returned unexpected error: %v", err)
				return
			}

			// Verify message was published
			publishedMsg := channel.GetLastPublishedMessage()
			if publishedMsg == nil {
				t.Errorf("No message was published")
				return
			}

			// Verify delivery mode
			if publishedMsg.Publishing.DeliveryMode != tt.expectedDM {
				t.Errorf("DeliveryMode = %d, want %d", publishedMsg.Publishing.DeliveryMode, tt.expectedDM)
			}

			// Verify headers - check that expected headers are present
			for expectedKey, expectedValue := range tt.expectedHdrs {
				actualValue, exists := publishedMsg.Publishing.Headers[expectedKey]
				if !exists {
					t.Errorf("Header %v not found in published message", expectedKey)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("Header[%v] = %v, want %v", expectedKey, actualValue, expectedValue)
				}
			}

			// Verify no unexpected headers (except tracing headers)
			for actualKey := range publishedMsg.Publishing.Headers {
				if _, expected := tt.expectedHdrs[actualKey]; !expected {
					// Allow tracing headers (they start with specific prefixes)
					if actualKey != "traceparent" && actualKey != "tracestate" {
						t.Logf("Found unexpected header: %v", actualKey)
					}
				}
			}

			// Verify other publishing parameters are set correctly
			if publishedMsg.Exchange != "test-exchange" {
				t.Errorf("Exchange = %v, want test-exchange", publishedMsg.Exchange)
			}
			if publishedMsg.Key != "test.key" {
				t.Errorf("Key = %v, want test.key", publishedMsg.Key)
			}
			if publishedMsg.Publishing.ContentType != JSONContentType {
				t.Errorf("ContentType = %v, want %v", publishedMsg.Publishing.ContentType, JSONContentType)
			}
			if publishedMsg.Publishing.AppId != "test-app" {
				t.Errorf("AppId = %v, want test-app", publishedMsg.Publishing.AppId)
			}
		})
	}
}

func TestPublisher_PublishDeadlineWithOptions(t *testing.T) {
	manager := NewMockConnectionManager()
	channel := NewMockAMQPChannel()
	manager.SetChannel(channel)

	publisher := NewPublisher("test-app", manager)

	// Test with options
	options := NewOption().WithDeliveryMode(amqp091.Persistent).WithHeaders(map[string]any{"priority": 10, "retry-enabled": true}).Build()

	err := publisher.PublishDeadline(
		context.Background(),
		"test-exchange",
		"test.deadline",
		TestMessage{ID: "deadline-test", Content: "deadline message"},
		options...,
	)

	// Expect error due to mock not having real broker confirmation
	if err == nil {
		t.Error("PublishDeadline should return error when no real confirmation available")
		return
	}

	t.Logf("PublishDeadline correctly returned error: %v", err)

	// Verify the options were processed by checking the published message
	publishedMsg := channel.GetLastPublishedMessage()
	if publishedMsg == nil {
		t.Errorf("No message was published")
		return
	}

	if publishedMsg.Publishing.DeliveryMode != amqp091.Persistent {
		t.Errorf("DeliveryMode = %d, want %d", publishedMsg.Publishing.DeliveryMode, amqp091.Persistent)
	}

	if priority, exists := publishedMsg.Publishing.Headers["priority"]; !exists || priority != 10 {
		t.Errorf("Header priority = %v, want 10", priority)
	}

	if retryEnabled, exists := publishedMsg.Publishing.Headers["retry-enabled"]; !exists || retryEnabled != true {
		t.Errorf("Header retry-enabled = %v, want true", retryEnabled)
	}
}

func TestPublisher_PublishQueueWithOptions(t *testing.T) {
	manager := NewMockConnectionManager()
	channel := NewMockAMQPChannel()
	manager.SetChannel(channel)

	publisher := NewPublisher("test-app", manager)

	// Test queue publishing with options
	options := append(
		NewOption().WithDeliveryMode(amqp091.Transient).Build(),
		NewOption().WithHeaders(map[string]any{"queue-specific": "value", "timestamp": 1234567890}).Build()...,
	)

	err := publisher.PublishQueue(
		context.Background(),
		"test-queue",
		TestMessage{ID: "queue-test", Content: "queue message"},
		options...,
	)

	if err != nil {
		t.Errorf("PublishQueue returned unexpected error: %v", err)
		return
	}

	// Verify the message was published to the queue (empty exchange)
	publishedMsg := channel.GetLastPublishedMessage()
	if publishedMsg == nil {
		t.Errorf("No message was published")
		return
	}

	if publishedMsg.Exchange != "" {
		t.Errorf("Exchange = %v, want empty string for queue publishing", publishedMsg.Exchange)
	}

	if publishedMsg.Key != "test-queue" {
		t.Errorf("Key = %v, want test-queue", publishedMsg.Key)
	}

	if publishedMsg.Publishing.DeliveryMode != amqp091.Transient {
		t.Errorf("DeliveryMode = %d, want %d", publishedMsg.Publishing.DeliveryMode, amqp091.Transient)
	}

	if queueSpecific, exists := publishedMsg.Publishing.Headers["queue-specific"]; !exists || queueSpecific != "value" {
		t.Errorf("Header queue-specific = %v, want 'value'", queueSpecific)
	}
}

func TestPublisher_PublishQueueDeadlineWithOptions(t *testing.T) {
	manager := NewMockConnectionManager()
	channel := NewMockAMQPChannel()
	manager.SetChannel(channel)

	publisher := NewPublisher("test-app", manager)

	// Test queue deadline publishing with options
	options := NewOption().WithDeliveryMode(amqp091.Persistent).WithHeaders(map[string]any{"deadline-queue": true, "timeout": "1s"}).Build()

	err := publisher.PublishQueueDeadline(
		context.Background(),
		"deadline-queue",
		TestMessage{ID: "queue-deadline-test", Content: "queue deadline message"},
		options...,
	)

	if err != nil {
		t.Errorf("PublishQueueDeadline returned unexpected error: %v", err)
		return
	}

	// Verify the message was published correctly
	publishedMsg := channel.GetLastPublishedMessage()
	if publishedMsg == nil {
		t.Errorf("No message was published")
		return
	}

	if publishedMsg.Exchange != "" {
		t.Errorf("Exchange = %v, want empty string for queue publishing", publishedMsg.Exchange)
	}

	if publishedMsg.Key != "deadline-queue" {
		t.Errorf("Key = %v, want deadline-queue", publishedMsg.Key)
	}

	if publishedMsg.Publishing.DeliveryMode != amqp091.Persistent {
		t.Errorf("DeliveryMode = %d, want %d", publishedMsg.Publishing.DeliveryMode, amqp091.Persistent)
	}

	if deadlineQueue, exists := publishedMsg.Publishing.Headers["deadline-queue"]; !exists || deadlineQueue != true {
		t.Errorf("Header deadline-queue = %v, want true", deadlineQueue)
	}
}

func TestPublisher_OptionsLookup(t *testing.T) {
	// Test the internal optionsLookup method behavior through published messages
	tests := []struct {
		name               string
		options            []*Option
		expectedDelivery   uint8
		expectedHeaderKeys []string
	}{
		{
			name:               "nil options",
			options:            nil,
			expectedDelivery:   0,
			expectedHeaderKeys: []string{},
		},
		{
			name:               "empty options slice",
			options:            []*Option{},
			expectedDelivery:   0,
			expectedHeaderKeys: []string{},
		},
		{
			name: "multiple options with same type - last wins",
			options: []*Option{
				{Key: OptionDeliveryModeKey, Value: amqp091.Transient},
				{Key: OptionHeadersKey, Value: map[string]any{"first": "headers"}},
				{Key: OptionDeliveryModeKey, Value: amqp091.Persistent},             // This should win
				{Key: OptionHeadersKey, Value: map[string]any{"second": "headers"}}, // This should win
			},
			expectedDelivery:   amqp091.Persistent,
			expectedHeaderKeys: []string{"second"},
		},
		{
			name: "unknown option keys are ignored",
			options: []*Option{
				{Key: OptionKey("unknown-key"), Value: "unknown-value"},
				{Key: OptionDeliveryModeKey, Value: amqp091.Transient},
				{Key: OptionKey("another-unknown"), Value: 42},
			},
			expectedDelivery:   amqp091.Transient,
			expectedHeaderKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConnectionManager()
			channel := NewMockAMQPChannel()
			manager.SetChannel(channel)

			publisher := NewPublisher("test-app", manager)

			// Clear any previous messages
			channel.ClearPublishedMessages()

			err := publisher.Publish(
				context.Background(),
				"test-exchange",
				"test.key",
				TestMessage{ID: "123", Content: "test"},
				tt.options...,
			)

			if err != nil {
				t.Errorf("Publish returned unexpected error: %v", err)
				return
			}

			// Verify delivery mode
			publishedMsg := channel.GetLastPublishedMessage()
			if publishedMsg == nil {
				t.Errorf("No message was published")
				return
			}

			if publishedMsg.Publishing.DeliveryMode != tt.expectedDelivery {
				t.Errorf("DeliveryMode = %d, want %d", publishedMsg.Publishing.DeliveryMode, tt.expectedDelivery)
			}

			// Verify expected header keys are present
			for _, expectedKey := range tt.expectedHeaderKeys {
				if _, exists := publishedMsg.Publishing.Headers[expectedKey]; !exists {
					t.Errorf("Expected header key %v not found", expectedKey)
				}
			}
		})
	}
}
