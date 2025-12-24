// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"crypto/tls"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

// =============================================================================
// MockAMQPChannel - Mock implementation of AMQPChannel interface for testing
// =============================================================================

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
	if m.publishError != nil {
		return nil, m.publishError
	}

	// Capture the published message for verification
	m.publishedMessages = append(m.publishedMessages, PublishedMessage{
		Exchange:   exchange,
		Key:        key,
		Mandatory:  mandatory,
		Immediate:  immediate,
		Publishing: msg,
	})

	// Return the stored deferredConfirmation or an error if none is set
	if m.deferredConfirmation == nil {
		return nil, errors.New("no deferred confirmation available - mock broker not running")
	}
	return m.deferredConfirmation, nil
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

func (m *MockAMQPChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	return nil
}

// =============================================================================
// MockRMQConnection - Mock implementation of RMQConnection interface for testing
// =============================================================================

// MockRMQConnection is a mock implementation of RMQConnection interface for testing
type MockRMQConnection struct {
	channels        []*amqp.Channel
	connectionState tls.ConnectionState
	closed          bool
	closeError      error
	channelError    error
	notifyChannels  []chan *amqp.Error
}

func NewMockRMQConnection() *MockRMQConnection {
	return &MockRMQConnection{
		channels:       make([]*amqp.Channel, 0),
		closed:         false,
		notifyChannels: make([]chan *amqp.Error, 0),
	}
}

func (m *MockRMQConnection) Channel() (*amqp.Channel, error) {
	if m.channelError != nil {
		return nil, m.channelError
	}
	// In a real test, we'd return a mock channel, but for interface testing we just return nil
	channel := &amqp.Channel{}
	m.channels = append(m.channels, channel)
	return channel, nil
}

func (m *MockRMQConnection) ConnectionState() tls.ConnectionState {
	return m.connectionState
}

func (m *MockRMQConnection) IsClosed() bool {
	return m.closed
}

func (m *MockRMQConnection) Close() error {
	m.closed = true
	return m.closeError
}

func (m *MockRMQConnection) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	m.notifyChannels = append(m.notifyChannels, receiver)
	return receiver
}

// Helper methods for testing
func (m *MockRMQConnection) SetChannelError(err error) {
	m.channelError = err
}

func (m *MockRMQConnection) SetCloseError(err error) {
	m.closeError = err
}

func (m *MockRMQConnection) SetConnectionState(state tls.ConnectionState) {
	m.connectionState = state
}

func (m *MockRMQConnection) TriggerClose(err *amqp.Error) {
	for _, ch := range m.notifyChannels {
		select {
		case ch <- err:
		default:
		}
	}
	m.closed = true
}

// =============================================================================
// MockConnectionManager - Mock implementation of ConnectionManager interface for testing
// =============================================================================

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

// =============================================================================
// MockAcknowledger - Mock implementation of acknowledger interface for testing
// =============================================================================

// MockAcknowledger is a mock implementation of acknowledger interface for testing
type MockAcknowledger struct {
	ackFunc  func(multiple bool) error
	nackFunc func(multiple, requeue bool) error
}

func (m *MockAcknowledger) Ack(tag uint64, multiple bool) error {
	if m.ackFunc != nil {
		return m.ackFunc(multiple)
	}
	return nil
}

func (m *MockAcknowledger) Nack(tag uint64, multiple, requeue bool) error {
	if m.nackFunc != nil {
		return m.nackFunc(multiple, requeue)
	}
	return nil
}

func (m *MockAcknowledger) Reject(tag uint64, requeue bool) error {
	if m.nackFunc != nil {
		return m.nackFunc(false, requeue)
	}
	return nil
}
