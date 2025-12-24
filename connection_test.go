// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"crypto/tls"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestRMQConnection_Interface(t *testing.T) {
	// Test that MockRMQConnection implements RMQConnection interface
	var conn RMQConnection = NewMockRMQConnection()

	// Test Channel method
	channel, err := conn.Channel()
	if err != nil {
		t.Errorf("Channel() returned unexpected error: %v", err)
	}
	if channel == nil {
		t.Error("Channel() returned nil channel")
	}

	// Test ConnectionState method
	state := conn.ConnectionState()
	// Just verify it doesn't panic and returns a value
	_ = state

	// Test IsClosed method
	if conn.IsClosed() {
		t.Error("IsClosed() should return false for new connection")
	}

	// Test Close method
	err = conn.Close()
	if err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}

	// Test IsClosed after close
	if !conn.IsClosed() {
		t.Error("IsClosed() should return true after Close()")
	}

	// Test NotifyClose method
	notifyCh := make(chan *amqp.Error, 1)
	returnedCh := conn.NotifyClose(notifyCh)
	if returnedCh != notifyCh {
		t.Error("NotifyClose() should return the same channel passed to it")
	}
}

func TestMockRMQConnection_Channel(t *testing.T) {
	conn := NewMockRMQConnection()

	// Test successful channel creation
	channel, err := conn.Channel()
	if err != nil {
		t.Errorf("Channel() returned unexpected error: %v", err)
	}
	if channel == nil {
		t.Error("Channel() returned nil channel")
	}

	// Test channel creation with error
	expectedError := &amqp.Error{Code: 404, Reason: "channel error"}
	conn.SetChannelError(expectedError)
	channel, err = conn.Channel()
	if err != expectedError {
		t.Errorf("Channel() with error should return expected error, got %v", err)
	}
	if channel != nil {
		t.Error("Channel() with error should return nil channel")
	}
}

func TestMockRMQConnection_ConnectionState(t *testing.T) {
	conn := NewMockRMQConnection()

	// Test default connection state
	state := conn.ConnectionState()
	if state.Version != 0 {
		t.Errorf("Default ConnectionState version should be 0, got %d", state.Version)
	}

	// Test setting connection state
	expectedState := tls.ConnectionState{
		Version: tls.VersionTLS12,
	}
	conn.SetConnectionState(expectedState)
	state = conn.ConnectionState()
	if state.Version != tls.VersionTLS12 {
		t.Errorf("ConnectionState version should be %d, got %d", tls.VersionTLS12, state.Version)
	}
}

func TestMockRMQConnection_IsClosed(t *testing.T) {
	conn := NewMockRMQConnection()

	// Test initial state
	if conn.IsClosed() {
		t.Error("New connection should not be closed")
	}

	// Test after close
	err := conn.Close()
	if err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
	if !conn.IsClosed() {
		t.Error("Connection should be closed after Close()")
	}
}

func TestMockRMQConnection_Close(t *testing.T) {
	conn := NewMockRMQConnection()

	// Test successful close
	err := conn.Close()
	if err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
	if !conn.IsClosed() {
		t.Error("Connection should be marked as closed")
	}

	// Test close with error
	conn = NewMockRMQConnection()
	expectedError := &amqp.Error{Code: 500, Reason: "close error"}
	conn.SetCloseError(expectedError)
	err = conn.Close()
	if err != expectedError {
		t.Errorf("Close() with error should return expected error, got %v", err)
	}
	if !conn.IsClosed() {
		t.Error("Connection should be marked as closed even with error")
	}
}

func TestMockRMQConnection_NotifyClose(t *testing.T) {
	conn := NewMockRMQConnection()

	// Test NotifyClose
	notifyCh := make(chan *amqp.Error, 1)
	returnedCh := conn.NotifyClose(notifyCh)
	if returnedCh != notifyCh {
		t.Error("NotifyClose() should return the same channel")
	}

	// Test triggering close notification
	expectedError := &amqp.Error{Code: 500, Reason: "connection closed"}
	conn.TriggerClose(expectedError)

	select {
	case receivedError := <-notifyCh:
		if receivedError != expectedError {
			t.Errorf("Received error %v, expected %v", receivedError, expectedError)
		}
	default:
		t.Error("Should have received close notification")
	}

	if !conn.IsClosed() {
		t.Error("Connection should be marked as closed after TriggerClose")
	}
}

func TestMockRMQConnection_MultipleChannels(t *testing.T) {
	conn := NewMockRMQConnection()

	// Create multiple channels
	numChannels := 3
	for i := 0; i < numChannels; i++ {
		channel, err := conn.Channel()
		if err != nil {
			t.Errorf("Channel() #%d returned error: %v", i, err)
		}
		if channel == nil {
			t.Errorf("Channel() #%d returned nil", i)
		}
	}

	// Verify channels were tracked
	if len(conn.channels) != numChannels {
		t.Errorf("Expected %d channels to be tracked, got %d", numChannels, len(conn.channels))
	}
}

func TestMockRMQConnection_MultipleNotifyChannels(t *testing.T) {
	conn := NewMockRMQConnection()

	// Register multiple notify channels
	numNotifyChannels := 3
	notifyChannels := make([]chan *amqp.Error, numNotifyChannels)
	for i := 0; i < numNotifyChannels; i++ {
		notifyChannels[i] = make(chan *amqp.Error, 1)
		conn.NotifyClose(notifyChannels[i])
	}

	// Trigger close and verify all channels receive notification
	expectedError := &amqp.Error{Code: 500, Reason: "connection closed"}
	conn.TriggerClose(expectedError)

	for i, ch := range notifyChannels {
		select {
		case receivedError := <-ch:
			if receivedError != expectedError {
				t.Errorf("Channel %d received error %v, expected %v", i, receivedError, expectedError)
			}
		default:
			t.Errorf("Channel %d should have received close notification", i)
		}
	}
}
