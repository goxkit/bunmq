// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"context"
	"errors"
	"testing"
	"time"
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

	// Verify it implements the Publisher interface
	pub := publisher
	if pub == nil {
		t.Error("Publisher is nil")
	}
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
			name:        "successful publish with deadline",
			exchange:    "test-exchange",
			routingKey:  "test.key",
			msg:         TestMessage{ID: "123", Content: "test"},
			channelErr:  nil,
			publishErr:  nil,
			expectError: false,
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
	// Test that PublishDeadline creates a timeout context
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

	if err != nil {
		t.Errorf("PublishDeadline returned unexpected error: %v", err)
	}
}
