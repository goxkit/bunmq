// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"context"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type (
	// ConnectionManager manages RabbitMQ connections and channels with automatic reconnection
	// It monitors channel and connection health using NotifyClose and NotifyCancel
	ConnectionManager interface {
		// GetConnection returns the current connection, ensuring it's healthy
		GetConnection() (RMQConnection, error)

		// GetChannel returns the current channel, ensuring it's healthy
		GetChannel() (AMQPChannel, error)

		GetConnectionString() string

		// Close gracefully closes the connection manager
		Close() error

		// IsHealthy checks if both connection and channel are healthy
		IsHealthy() bool

		// SetReconnectCallback sets a callback function that's called when reconnection occurs
		SetReconnectCallback(callback func(conn RMQConnection, ch AMQPChannel))
	}

	// connectionManager implements ConnectionManager with automatic reconnection capabilities
	connectionManager struct {
		connectionString  string
		conn              RMQConnection
		ch                AMQPChannel
		mu                sync.RWMutex
		closed            bool
		reconnectCallback func(RMQConnection, AMQPChannel)

		// Reconnection configuration
		maxReconnectAttempts int
		reconnectDelay       time.Duration
		reconnectBackoffMax  time.Duration

		// Channels for monitoring connection/channel health
		connCloseNotify chan *amqp.Error
		chCloseNotify   chan *amqp.Error
		chCancelNotify  chan string

		// Context for cancellation
		ctx    context.Context
		cancel context.CancelFunc
	}

	// ReconnectionConfig holds configuration for reconnection behavior
	ReconnectionConfig struct {
		MaxAttempts   int           // Maximum reconnection attempts (0 = infinite)
		InitialDelay  time.Duration // Initial delay between reconnection attempts
		BackoffMax    time.Duration // Maximum delay between attempts
		BackoffFactor float64       // Exponential backoff factor
	}
)

// DefaultReconnectionConfig provides sensible defaults for reconnection behavior
var DefaultReconnectionConfig = ReconnectionConfig{
	MaxAttempts:   0,               // Infinite attempts
	InitialDelay:  time.Second * 2, // Start with 2 second delay
	BackoffMax:    time.Minute * 5, // Maximum 5 minute delay
	BackoffFactor: 1.5,             // 1.5x exponential backoff
}

// NewConnectionManager creates a new connection manager with automatic reconnection
func NewConnectionManager(connectionString string, config ...ReconnectionConfig) (ConnectionManager, error) {
	cfg := DefaultReconnectionConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	ctx, cancel := context.WithCancel(context.Background())

	cm := &connectionManager{
		connectionString:     connectionString,
		maxReconnectAttempts: cfg.MaxAttempts,
		reconnectDelay:       cfg.InitialDelay,
		reconnectBackoffMax:  cfg.BackoffMax,
		ctx:                  ctx,
		cancel:               cancel,
	}

	// Establish initial connection
	if err := cm.connect(); err != nil {
		cancel()
		return nil, err
	}

	// Start monitoring goroutine
	go cm.monitor()

	return cm, nil
}

// connect establishes connection and channel, setting up health monitoring
func (cm *connectionManager) connect() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	logrus.Info("bunmq establishing connection...")

	// Create connection
	conn, ch, err := NewConnection(cm.connectionString)
	if err != nil {
		logrus.WithError(err).Error("bunmq failed to establish connection")
		return err
	}

	// Set up notification channels for monitoring
	cm.setupNotificationChannels(conn, ch)

	cm.conn = conn
	cm.ch = ch

	logrus.Info("bunmq connection established successfully")

	// Call reconnect callback if set
	if cm.reconnectCallback != nil {
		cm.reconnectCallback(conn, ch)
	}

	return nil
}

// setupNotificationChannels configures NotifyClose and NotifyCancel monitoring
func (cm *connectionManager) setupNotificationChannels(conn RMQConnection, ch AMQPChannel) {
	// Monitor connection close events
	if amqpConn, ok := conn.(*amqp.Connection); ok {
		cm.connCloseNotify = make(chan *amqp.Error, 1)
		amqpConn.NotifyClose(cm.connCloseNotify)
	}

	// Monitor channel close and cancel events
	if amqpCh, ok := ch.(*amqp.Channel); ok {
		cm.chCloseNotify = make(chan *amqp.Error, 1)
		cm.chCancelNotify = make(chan string, 1)

		amqpCh.NotifyClose(cm.chCloseNotify)
		amqpCh.NotifyCancel(cm.chCancelNotify)
	}
}

// monitor runs in a goroutine to watch for connection/channel issues
func (cm *connectionManager) monitor() {
	for {
		select {
		case <-cm.ctx.Done():
			logrus.Info("bunmq connection manager stopped")
			return

		case err := <-cm.connCloseNotify:
			if err != nil {
				logrus.WithError(err).Warn("bunmq connection closed unexpectedly")
				cm.handleConnectionFailure()
			}

		case err := <-cm.chCloseNotify:
			if err != nil {
				logrus.WithError(err).Warn("bunmq channel closed unexpectedly")
				cm.handleChannelFailure()
			}

		case consumerTag := <-cm.chCancelNotify:
			logrus.WithField("consumerTag", consumerTag).Warn("bunmq consumer cancelled")
			// For consumer cancellation, we typically want to recreate the channel
			cm.handleChannelFailure()
		}
	}
}

// handleChannelFailure attempts to recreate the channel while keeping the connection
func (cm *connectionManager) handleChannelFailure() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.closed {
		return
	}

	logrus.Info("bunmq attempting to recreate channel...")

	// Check if connection is still healthy
	if cm.conn != nil && !cm.conn.IsClosed() {
		// Try to create a new channel on the existing connection
		newCh, err := cm.conn.Channel()
		if err != nil {
			logrus.WithError(err).Error("bunmq failed to recreate channel, will reconnect completely")
			cm.handleConnectionFailure()
			return
		}

		// Set up monitoring for the new channel
		cm.setupChannelNotifications(newCh)
		cm.ch = newCh

		logrus.Info("bunmq channel recreated successfully")

		// Call reconnect callback if set
		if cm.reconnectCallback != nil {
			cm.reconnectCallback(cm.conn, cm.ch)
		}
	} else {
		// Connection is also bad, do full reconnection
		cm.handleConnectionFailure()
	}
}

// setupChannelNotifications sets up only channel-specific notifications
func (cm *connectionManager) setupChannelNotifications(ch *amqp.Channel) {
	cm.chCloseNotify = make(chan *amqp.Error, 1)
	cm.chCancelNotify = make(chan string, 1)

	ch.NotifyClose(cm.chCloseNotify)
	ch.NotifyCancel(cm.chCancelNotify)
}

// handleConnectionFailure attempts full reconnection with exponential backoff
func (cm *connectionManager) handleConnectionFailure() {
	go cm.reconnectWithBackoff()
}

// reconnectWithBackoff implements exponential backoff reconnection strategy
func (cm *connectionManager) reconnectWithBackoff() {
	attempt := 0
	delay := cm.reconnectDelay

	for {
		select {
		case <-cm.ctx.Done():
			return
		default:
		}

		attempt++

		// Check if we've exceeded max attempts
		if cm.maxReconnectAttempts > 0 && attempt > cm.maxReconnectAttempts {
			logrus.Errorf("bunmq exceeded maximum reconnection attempts (%d)", cm.maxReconnectAttempts)
			return
		}

		logrus.WithFields(logrus.Fields{
			"attempt": attempt,
			"delay":   delay,
		}).Info("bunmq attempting reconnection...")

		// Wait before attempting reconnection
		select {
		case <-cm.ctx.Done():
			return
		case <-time.After(delay):
		}

		// Attempt reconnection
		if err := cm.connect(); err != nil {
			logrus.WithError(err).WithField("attempt", attempt).Error("bunmq reconnection failed")

			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * 1.5)
			if delay > cm.reconnectBackoffMax {
				delay = cm.reconnectBackoffMax
			}

			continue
		}

		logrus.WithField("attempt", attempt).Info("bunmq reconnection successful")
		return
	}
}

func (cm *connectionManager) GetConnectionString() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.closed {
		return ""
	}

	return cm.connectionString
}

// GetConnection returns the current connection, ensuring it's healthy
func (cm *connectionManager) GetConnection() (RMQConnection, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.closed {
		return nil, NewBunMQError("connection manager is closed")
	}

	if cm.conn == nil || cm.conn.IsClosed() {
		return nil, NewBunMQError("connection is not available")
	}

	return cm.conn, nil
}

// GetChannel returns the current channel, ensuring it's healthy
func (cm *connectionManager) GetChannel() (AMQPChannel, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.closed {
		return nil, NewBunMQError("connection manager is closed")
	}

	if cm.ch == nil || cm.ch.IsClosed() {
		return nil, NewBunMQError("channel is not available")
	}

	return cm.ch, nil
}

// IsHealthy checks if both connection and channel are healthy
func (cm *connectionManager) IsHealthy() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.closed {
		return false
	}

	return cm.conn != nil && !cm.conn.IsClosed() &&
		cm.ch != nil && !cm.ch.IsClosed()
}

// SetReconnectCallback sets a callback function that's called when reconnection occurs
func (cm *connectionManager) SetReconnectCallback(callback func(conn RMQConnection, ch AMQPChannel)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.reconnectCallback = callback
}

// Close gracefully closes the connection manager
func (cm *connectionManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.closed {
		return nil
	}

	cm.closed = true
	cm.cancel()

	var err error
	if cm.ch != nil && !cm.ch.IsClosed() {
		if closeErr := cm.ch.Close(); closeErr != nil {
			logrus.WithError(closeErr).Error("bunmq error closing channel")
			err = closeErr
		}
	}

	if cm.conn != nil && !cm.conn.IsClosed() {
		if closeErr := cm.conn.Close(); closeErr != nil {
			logrus.WithError(closeErr).Error("bunmq error closing connection")
			err = closeErr
		}
	}

	logrus.Info("bunmq connection manager closed")
	return err
}
