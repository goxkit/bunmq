// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestAMQPChannel_Interface(t *testing.T) {
	// Test that MockAMQPChannel implements AMQPChannel interface
	var channel AMQPChannel = NewMockAMQPChannel()

	// Test ExchangeDeclare
	err := channel.ExchangeDeclare("test-exchange", "direct", true, false, false, false, nil)
	if err != nil {
		t.Errorf("ExchangeDeclare() returned unexpected error: %v", err)
	}

	// Test ExchangeBind
	err = channel.ExchangeBind("dest", "key", "source", false, nil)
	if err != nil {
		t.Errorf("ExchangeBind() returned unexpected error: %v", err)
	}

	// Test QueueDeclare
	queue, err := channel.QueueDeclare("test-queue", true, false, false, false, nil)
	if err != nil {
		t.Errorf("QueueDeclare() returned unexpected error: %v", err)
	}
	if queue.Name != "test-queue" {
		t.Errorf("QueueDeclare() returned queue name %v, want test-queue", queue.Name)
	}

	// Test QueueBind
	err = channel.QueueBind("queue", "key", "exchange", false, nil)
	if err != nil {
		t.Errorf("QueueBind() returned unexpected error: %v", err)
	}

	// Test Consume
	deliveries, err := channel.Consume("queue", "consumer", true, false, false, false, nil)
	if err != nil {
		t.Errorf("Consume() returned unexpected error: %v", err)
	}
	if deliveries == nil {
		t.Error("Consume() returned nil delivery channel")
	}

	// Test Publish
	msg := amqp.Publishing{Body: []byte("test")}
	err = channel.Publish("exchange", "key", false, false, msg)
	if err != nil {
		t.Errorf("Publish() returned unexpected error: %v", err)
	}

	// Test IsClosed
	if channel.IsClosed() {
		t.Error("IsClosed() should return false for new channel")
	}

	// Test Close
	err = channel.Close()
	if err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}

	// Test IsClosed after close
	if !channel.IsClosed() {
		t.Error("IsClosed() should return true after Close()")
	}

	// Test NotifyClose
	notifyCloseCh := make(chan *amqp.Error, 1)
	returnedCh := channel.NotifyClose(notifyCloseCh)
	if returnedCh != notifyCloseCh {
		t.Error("NotifyClose() should return the same channel passed to it")
	}

	// Test NotifyCancel
	notifyCancelCh := make(chan string, 1)
	returnedCancelCh := channel.NotifyCancel(notifyCancelCh)
	if returnedCancelCh != notifyCancelCh {
		t.Error("NotifyCancel() should return the same channel passed to it")
	}
}

func TestMockAMQPChannel_ExchangeDeclare(t *testing.T) {
	channel := NewMockAMQPChannel()

	// Test successful exchange declaration
	err := channel.ExchangeDeclare("orders", "direct", true, false, false, false, nil)
	if err != nil {
		t.Errorf("ExchangeDeclare() returned unexpected error: %v", err)
	}

	// Test exchange declaration with error
	expectedError := &amqp.Error{Code: 404, Reason: "exchange declare error"}
	channel.SetExchangeDeclareError(expectedError)
	err = channel.ExchangeDeclare("orders", "direct", true, false, false, false, nil)
	if err != expectedError {
		t.Errorf("ExchangeDeclare() with error should return expected error, got %v", err)
	}
}

func TestMockAMQPChannel_QueueDeclare(t *testing.T) {
	channel := NewMockAMQPChannel()

	// Test successful queue declaration
	queue, err := channel.QueueDeclare("orders", true, false, false, false, nil)
	if err != nil {
		t.Errorf("QueueDeclare() returned unexpected error: %v", err)
	}
	if queue.Name != "orders" {
		t.Errorf("QueueDeclare() returned queue name %v, want orders", queue.Name)
	}

	// Test queue declaration with error
	expectedError := &amqp.Error{Code: 404, Reason: "queue declare error"}
	channel.SetQueueDeclareError(expectedError)
	queue, err = channel.QueueDeclare("orders", true, false, false, false, nil)
	if err != expectedError {
		t.Errorf("QueueDeclare() with error should return expected error, got %v", err)
	}
	if queue.Name != "" {
		t.Errorf("QueueDeclare() with error should return empty queue, got %v", queue)
	}
}

func TestMockAMQPChannel_Consume(t *testing.T) {
	channel := NewMockAMQPChannel()

	// Test successful consume
	deliveries, err := channel.Consume("orders", "consumer1", true, false, false, false, nil)
	if err != nil {
		t.Errorf("Consume() returned unexpected error: %v", err)
	}
	if deliveries == nil {
		t.Error("Consume() should return delivery channel")
	}

	// Test consume with error
	expectedError := &amqp.Error{Code: 404, Reason: "consume error"}
	channel.SetConsumeError(expectedError)
	deliveries, err = channel.Consume("orders", "consumer1", true, false, false, false, nil)
	if err != expectedError {
		t.Errorf("Consume() with error should return expected error, got %v", err)
	}
	if deliveries != nil {
		t.Error("Consume() with error should return nil delivery channel")
	}
}

func TestMockAMQPChannel_Publish(t *testing.T) {
	channel := NewMockAMQPChannel()

	// Test successful publish
	msg := amqp.Publishing{Body: []byte("test message")}
	err := channel.Publish("orders", "order.created", false, false, msg)
	if err != nil {
		t.Errorf("Publish() returned unexpected error: %v", err)
	}

	// Test publish with error
	expectedError := &amqp.Error{Code: 404, Reason: "publish error"}
	channel.SetPublishError(expectedError)
	err = channel.Publish("orders", "order.created", false, false, msg)
	if err != expectedError {
		t.Errorf("Publish() with error should return expected error, got %v", err)
	}
}

func TestMockAMQPChannel_NotifyClose(t *testing.T) {
	channel := NewMockAMQPChannel()

	// Test NotifyClose
	notifyCloseCh := make(chan *amqp.Error, 1)
	returnedCh := channel.NotifyClose(notifyCloseCh)
	if returnedCh != notifyCloseCh {
		t.Error("NotifyClose() should return the same channel")
	}

	// Test triggering close notification
	expectedError := &amqp.Error{Code: 500, Reason: "channel closed"}
	channel.TriggerClose(expectedError)

	select {
	case receivedError := <-notifyCloseCh:
		if receivedError != expectedError {
			t.Errorf("Received error %v, expected %v", receivedError, expectedError)
		}
	default:
		t.Error("Should have received close notification")
	}

	if !channel.IsClosed() {
		t.Error("Channel should be marked as closed after TriggerClose")
	}
}

func TestMockAMQPChannel_NotifyCancel(t *testing.T) {
	channel := NewMockAMQPChannel()

	// Test NotifyCancel
	notifyCancelCh := make(chan string, 1)
	returnedCh := channel.NotifyCancel(notifyCancelCh)
	if returnedCh != notifyCancelCh {
		t.Error("NotifyCancel() should return the same channel")
	}

	// Test triggering cancel notification
	expectedConsumer := "consumer-123"
	channel.TriggerCancel(expectedConsumer)

	select {
	case receivedConsumer := <-notifyCancelCh:
		if receivedConsumer != expectedConsumer {
			t.Errorf("Received consumer %v, expected %v", receivedConsumer, expectedConsumer)
		}
	default:
		t.Error("Should have received cancel notification")
	}
}

func TestNewConnection(t *testing.T) {
	originalDial := dial
	defer func() { dial = originalDial }()

	tests := []struct {
		name             string
		appName          string
		connectionString string
		dialError        error
		channelError     error
		expectError      bool
	}{
		{
			name:             "successful connection",
			appName:          "test-app",
			connectionString: "amqp://guest:guest@localhost:5672/",
			dialError:        nil,
			channelError:     nil,
			expectError:      false,
		},
		{
			name:             "dial error",
			appName:          "test-app",
			connectionString: "amqp://invalid",
			dialError:        errors.New("dial error"),
			channelError:     nil,
			expectError:      true,
		},
		{
			name:             "channel error",
			appName:          "test-app",
			connectionString: "amqp://guest:guest@localhost:5672/",
			dialError:        nil,
			channelError:     errors.New("channel error"),
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the dial function
			dial = func(appName, connectionString string) (RMQConnection, error) {
				if tt.dialError != nil {
					return nil, tt.dialError
				}

				mockConn := NewMockRMQConnection()
				if tt.channelError != nil {
					mockConn.SetChannelError(tt.channelError)
				}
				return mockConn, nil
			}

			conn, channel, err := NewConnection(tt.appName, tt.connectionString)

			if tt.expectError {
				if err == nil {
					t.Error("NewConnection() should return error")
				}
				if conn != nil || channel != nil {
					t.Error("NewConnection() with error should return nil connection and channel")
				}
			} else {
				if err != nil {
					t.Errorf("NewConnection() returned unexpected error: %v", err)
				}
				if conn == nil {
					t.Error("NewConnection() should return connection")
				}
				if channel == nil {
					t.Error("NewConnection() should return channel")
				}
			}
		})
	}
}

func TestDial_Variable(t *testing.T) {
	// Test that dial variable exists and is a function
	if dial == nil {
		t.Fatal("dial variable is nil")
	}

	// Test that dial has the correct signature
	// We can't call it directly in tests without a real RabbitMQ instance,
	// but we can verify it exists and is callable

	// Store original dial function
	originalDial := dial
	defer func() { dial = originalDial }()

	// Test that we can replace dial for mocking
	mockCalled := false
	dial = func(appName, connectionString string) (RMQConnection, error) {
		mockCalled = true
		return NewMockRMQConnection(), nil
	}

	// Use NewConnection to verify our mock is called
	_, _, err := NewConnection("test", "test-connection")
	if err != nil {
		t.Errorf("NewConnection() with mocked dial returned error: %v", err)
	}
	if !mockCalled {
		t.Error("Mocked dial function was not called")
	}
}
