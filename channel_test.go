// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MockAMQPChannel is a mock implementation of AMQPChannel interface for testing
type MockAMQPChannel struct {
	exchangeDeclareError error
	exchangeBindError    error
	queueDeclareError    error
	queueDeclareQueue    amqp.Queue
	queueBindError       error
	consumeError         error
	consumeChannel       <-chan amqp.Delivery
	deferredConfirmation *amqp.DeferredConfirmation
	publishError         error
	closed               bool
	closeError           error
	notifyCloseChannels  []chan *amqp.Error
	notifyCancelChannels []chan string
	// declaredArgs captures the args passed to QueueDeclare for tests
	declaredArgs map[string]amqp.Table
	// publishedMessages captures published messages for verification
	publishedMessages []PublishedMessage
}

// PublishedMessage captures the details of a published message
type PublishedMessage struct {
	Exchange   string
	Key        string
	Mandatory  bool
	Immediate  bool
	Publishing amqp.Publishing
}

func NewMockAMQPChannel() *MockAMQPChannel {
	// Create a default delivery channel for testing
	deliveryChannel := make(chan amqp.Delivery)
	close(deliveryChannel) // Close it immediately to simulate an empty channel

	return &MockAMQPChannel{
		closed:               false,
		notifyCloseChannels:  make([]chan *amqp.Error, 0),
		notifyCancelChannels: make([]chan string, 0),
		queueDeclareQueue:    amqp.Queue{Name: "test-queue", Messages: 0, Consumers: 0},
		consumeChannel:       deliveryChannel,
		publishedMessages:    make([]PublishedMessage, 0),
	}
}

func (m *MockAMQPChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return m.exchangeDeclareError
}

func (m *MockAMQPChannel) ExchangeBind(destination, key, source string, noWait bool, args amqp.Table) error {
	return m.exchangeBindError
}

func (m *MockAMQPChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	if m.queueDeclareError != nil {
		return amqp.Queue{}, m.queueDeclareError
	}
	if m.declaredArgs == nil {
		m.declaredArgs = map[string]amqp.Table{}
	}
	// copy args map to avoid mutation issues from caller
	copied := amqp.Table{}
	for k, v := range args {
		copied[k] = v
	}
	m.declaredArgs[name] = copied

	queue := m.queueDeclareQueue
	queue.Name = name
	return queue, nil
}

func (m *MockAMQPChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return m.queueBindError
}

func (m *MockAMQPChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	if m.consumeError != nil {
		return nil, m.consumeError
	}
	return m.consumeChannel, nil
}

func (m *MockAMQPChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	// Capture published message for verification in tests
	m.publishedMessages = append(m.publishedMessages, PublishedMessage{
		Exchange:   exchange,
		Key:        key,
		Mandatory:  mandatory,
		Immediate:  immediate,
		Publishing: msg,
	})
	return m.publishError
}

func (m *MockAMQPChannel) PublishWithDeferredConfirm(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) (*amqp.DeferredConfirmation, error) {
	return m.deferredConfirmation, m.publishError
}

func (m *MockAMQPChannel) IsClosed() bool {
	return m.closed
}

func (m *MockAMQPChannel) Close() error {
	m.closed = true
	return m.closeError
}

func (m *MockAMQPChannel) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	m.notifyCloseChannels = append(m.notifyCloseChannels, receiver)
	return receiver
}

func (m *MockAMQPChannel) NotifyCancel(receiver chan string) chan string {
	m.notifyCancelChannels = append(m.notifyCancelChannels, receiver)
	return receiver
}

// Helper methods for testing
func (m *MockAMQPChannel) SetExchangeDeclareError(err error) {
	m.exchangeDeclareError = err
}

func (m *MockAMQPChannel) SetExchangeBindError(err error) {
	m.exchangeBindError = err
}

func (m *MockAMQPChannel) SetQueueDeclareError(err error) {
	m.queueDeclareError = err
}

func (m *MockAMQPChannel) SetQueueDeclareQueue(queue amqp.Queue) {
	m.queueDeclareQueue = queue
}

func (m *MockAMQPChannel) SetQueueBindError(err error) {
	m.queueBindError = err
}

func (m *MockAMQPChannel) SetConsumeError(err error) {
	m.consumeError = err
}

func (m *MockAMQPChannel) SetConsumeChannel(ch <-chan amqp.Delivery) {
	m.consumeChannel = ch
}

func (m *MockAMQPChannel) SetDeferredConfirmation(dc *amqp.DeferredConfirmation) {
	m.deferredConfirmation = dc
}

func (m *MockAMQPChannel) SetPublishError(err error) {
	m.publishError = err
}

// GetPublishedMessages returns all captured published messages
func (m *MockAMQPChannel) GetPublishedMessages() []PublishedMessage {
	return m.publishedMessages
}

// GetLastPublishedMessage returns the last published message, or nil if none
func (m *MockAMQPChannel) GetLastPublishedMessage() *PublishedMessage {
	if len(m.publishedMessages) == 0 {
		return nil
	}
	return &m.publishedMessages[len(m.publishedMessages)-1]
}

// ClearPublishedMessages clears the published messages history
func (m *MockAMQPChannel) ClearPublishedMessages() {
	m.publishedMessages = make([]PublishedMessage, 0)
}

func (m *MockAMQPChannel) SetCloseError(err error) {
	m.closeError = err
}

func (m *MockAMQPChannel) TriggerClose(err *amqp.Error) {
	for _, ch := range m.notifyCloseChannels {
		select {
		case ch <- err:
		default:
		}
	}
	m.closed = true
}

func (m *MockAMQPChannel) TriggerCancel(consumer string) {
	for _, ch := range m.notifyCancelChannels {
		select {
		case ch <- consumer:
		default:
		}
	}
}

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
