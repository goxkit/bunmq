// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Test helper to create a working connection manager with mocks
func createMockConnectionManager(t *testing.T, config ...ReconnectionConfig) (*connectionManager, *MockRMQConnection, *MockAMQPChannel) {
	cfg := DefaultReconnectionConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	mockConn := NewMockRMQConnection()
	mockCh := NewMockAMQPChannel()

	cm := &connectionManager{
		connectionString:     "amqp://test",
		appName:              "test-app",
		maxReconnectAttempts: cfg.MaxAttempts,
		reconnectDelay:       cfg.InitialDelay,
		reconnectBackoffMax:  cfg.BackoffMax,
		ctx:                  ctx,
		cancel:               cancel,
		conn:                 mockConn,
		ch:                   mockCh,
		closed:               false,
	}

	// Set up notification channels
	cm.connCloseNotify = make(chan *amqp.Error, 1)
	cm.chCloseNotify = make(chan *amqp.Error, 1)
	cm.chCancelNotify = make(chan string, 1)

	return cm, mockConn, mockCh
}

func TestNewConnectionManager_Configuration(t *testing.T) {
	// Test only the configuration aspects since connection creation is hard to mock
	tests := []struct {
		name   string
		config ReconnectionConfig
	}{
		{
			name:   "default configuration",
			config: DefaultReconnectionConfig,
		},
		{
			name: "custom configuration",
			config: ReconnectionConfig{
				MaxAttempts:   5,
				InitialDelay:  time.Second,
				BackoffMax:    time.Minute,
				BackoffFactor: 2.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since we can't easily mock NewConnection, we test the configuration handling
			// by creating a connection manager directly with mocks
			cm, _, _ := createMockConnectionManager(t, tt.config)

			if cm.maxReconnectAttempts != tt.config.MaxAttempts {
				t.Errorf("Expected MaxAttempts %d but got %d", tt.config.MaxAttempts, cm.maxReconnectAttempts)
			}
			if cm.reconnectDelay != tt.config.InitialDelay {
				t.Errorf("Expected InitialDelay %v but got %v", tt.config.InitialDelay, cm.reconnectDelay)
			}
			if cm.reconnectBackoffMax != tt.config.BackoffMax {
				t.Errorf("Expected BackoffMax %v but got %v", tt.config.BackoffMax, cm.reconnectBackoffMax)
			}
		})
	}
}

func TestConnectionManager_GetConnection(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*connectionManager, *MockRMQConnection)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful get connection",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection) {
				// Connection is healthy by default
			},
			expectError: false,
		},
		{
			name: "connection manager closed",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection) {
				cm.closed = true
			},
			expectError: true,
			errorMsg:    "connection manager is closed",
		},
		{
			name: "connection is closed",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection) {
				mockConn.Close()
			},
			expectError: true,
			errorMsg:    "connection is not available",
		},
		{
			name: "connection is nil",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection) {
				cm.conn = nil
			},
			expectError: true,
			errorMsg:    "connection is not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm, mockConn, _ := createMockConnectionManager(t)
			tt.setup(cm, mockConn)

			conn, err := cm.GetConnection()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if conn != nil {
					t.Error("Expected nil connection on error")
				}
				if tt.errorMsg != "" && err != nil {
					if err.Error() != tt.errorMsg {
						t.Errorf("Expected error message %q but got %q", tt.errorMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if conn == nil {
					t.Error("Expected connection but got nil")
				}
				if conn != mockConn {
					t.Error("Expected mock connection but got different connection")
				}
			}
		})
	}
}

func TestConnectionManager_GetChannel(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*connectionManager, *MockAMQPChannel)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful get channel",
			setup: func(cm *connectionManager, mockCh *MockAMQPChannel) {
				// Channel is healthy by default
			},
			expectError: false,
		},
		{
			name: "connection manager closed",
			setup: func(cm *connectionManager, mockCh *MockAMQPChannel) {
				cm.closed = true
			},
			expectError: true,
			errorMsg:    "connection manager is closed",
		},
		{
			name: "channel is closed",
			setup: func(cm *connectionManager, mockCh *MockAMQPChannel) {
				mockCh.Close()
			},
			expectError: true,
			errorMsg:    "channel is not available",
		},
		{
			name: "channel is nil",
			setup: func(cm *connectionManager, mockCh *MockAMQPChannel) {
				cm.ch = nil
			},
			expectError: true,
			errorMsg:    "channel is not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm, _, mockCh := createMockConnectionManager(t)
			tt.setup(cm, mockCh)

			ch, err := cm.GetChannel()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if ch != nil {
					t.Error("Expected nil channel on error")
				}
				if tt.errorMsg != "" && err != nil {
					if err.Error() != tt.errorMsg {
						t.Errorf("Expected error message %q but got %q", tt.errorMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if ch == nil {
					t.Error("Expected channel but got nil")
				}
				if ch != mockCh {
					t.Error("Expected mock channel but got different channel")
				}
			}
		})
	}
}

func TestConnectionManager_GetConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*connectionManager)
		expected string
	}{
		{
			name: "get connection string when healthy",
			setup: func(cm *connectionManager) {
				// Default healthy state
			},
			expected: "amqp://test",
		},
		{
			name: "get connection string when closed",
			setup: func(cm *connectionManager) {
				cm.closed = true
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm, _, _ := createMockConnectionManager(t)
			tt.setup(cm)

			result := cm.GetConnectionString()
			if result != tt.expected {
				t.Errorf("Expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestConnectionManager_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*connectionManager, *MockRMQConnection, *MockAMQPChannel)
		expected bool
	}{
		{
			name: "healthy connection and channel",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				// Default state is healthy
			},
			expected: true,
		},
		{
			name: "connection manager closed",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				cm.closed = true
			},
			expected: false,
		},
		{
			name: "connection is nil",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				cm.conn = nil
			},
			expected: false,
		},
		{
			name: "connection is closed",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				mockConn.Close()
			},
			expected: false,
		},
		{
			name: "channel is nil",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				cm.ch = nil
			},
			expected: false,
		},
		{
			name: "channel is closed",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				mockCh.Close()
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm, mockConn, mockCh := createMockConnectionManager(t)
			tt.setup(cm, mockConn, mockCh)

			result := cm.IsHealthy()
			if result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestConnectionManager_SetReconnectCallback(t *testing.T) {
	cm, mockConn, mockCh := createMockConnectionManager(t)

	callbackCalled := false
	var callbackConn RMQConnection
	var callbackCh AMQPChannel

	callback := func(conn RMQConnection, ch AMQPChannel) {
		callbackCalled = true
		callbackConn = conn
		callbackCh = ch
	}

	cm.SetReconnectCallback(callback)

	// Verify callback is set
	if cm.reconnectCallback == nil {
		t.Error("Expected callback to be set")
	}

	// Simulate calling the callback
	cm.reconnectCallback(mockConn, mockCh)

	if !callbackCalled {
		t.Error("Expected callback to be called")
	}
	if callbackConn != mockConn {
		t.Error("Expected callback to receive mock connection")
	}
	if callbackCh != mockCh {
		t.Error("Expected callback to receive mock channel")
	}
}

func TestConnectionManager_SetTopology(t *testing.T) {
	cm, _, _ := createMockConnectionManager(t)

	// Create a mock topology
	topology := &topology{
		appName:          "test-app",
		connectionString: "amqp://test",
	}

	cm.SetTopology(topology)

	if cm.topology != topology {
		t.Error("Expected topology to be set")
	}
}

func TestConnectionManager_Close(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*connectionManager, *MockRMQConnection, *MockAMQPChannel)
		expectErr bool
	}{
		{
			name: "successful close",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				// Default state
			},
			expectErr: false,
		},
		{
			name: "already closed",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				cm.closed = true
			},
			expectErr: false,
		},
		{
			name: "channel close error",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				mockCh.SetCloseError(errors.New("channel close error"))
			},
			expectErr: true,
		},
		{
			name: "connection close error",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				mockConn.SetCloseError(errors.New("connection close error"))
			},
			expectErr: true,
		},
		{
			name: "channel already closed",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				mockCh.Close()
			},
			expectErr: false,
		},
		{
			name: "connection already closed",
			setup: func(cm *connectionManager, mockConn *MockRMQConnection, mockCh *MockAMQPChannel) {
				mockConn.Close()
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm, mockConn, mockCh := createMockConnectionManager(t)
			tt.setup(cm, mockConn, mockCh)

			err := cm.Close()

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			// Verify the manager is marked as closed (except for "already closed" test)
			if tt.name != "already closed" && !cm.closed {
				t.Error("Expected connection manager to be marked as closed")
			}
		})
	}
}

func TestConnectionManager_ReconnectionStrategy(t *testing.T) {
	t.Run("channel failure triggers channel recreation", func(t *testing.T) {
		cm, _, _ := createMockConnectionManager(t)

		// Start monitoring
		go cm.monitor()

		// Trigger channel failure
		select {
		case cm.chCloseNotify <- &amqp.Error{Code: 505, Reason: "channel closed"}:
		default:
		}

		// Give some time for the handler to process
		time.Sleep(100 * time.Millisecond)

		// The test mainly verifies the flow doesn't panic and handles the event
	})

	t.Run("connection failure triggers full reconnection", func(t *testing.T) {
		config := ReconnectionConfig{
			MaxAttempts:   1, // Limit attempts for testing
			InitialDelay:  10 * time.Millisecond,
			BackoffMax:    100 * time.Millisecond,
			BackoffFactor: 1.5,
		}

		cm, _, _ := createMockConnectionManager(t, config)

		// Start monitoring
		go cm.monitor()

		// Trigger connection failure
		select {
		case cm.connCloseNotify <- &amqp.Error{Code: 320, Reason: "connection forced"}:
		default:
		}

		// Give some time for the handler to process
		time.Sleep(200 * time.Millisecond)

		// The test mainly verifies the flow doesn't panic and handles the event
	})

	t.Run("consumer cancellation triggers channel recreation", func(t *testing.T) {
		cm, _, _ := createMockConnectionManager(t)

		// Start monitoring
		go cm.monitor()

		// Trigger consumer cancellation
		select {
		case cm.chCancelNotify <- "test-consumer-tag":
		default:
		}

		// Give some time for the handler to process
		time.Sleep(100 * time.Millisecond)

		// The test mainly verifies the flow doesn't panic and handles the event
	})
}

func TestConnectionManager_ReconnectionConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config ReconnectionConfig
	}{
		{
			name:   "default configuration",
			config: DefaultReconnectionConfig,
		},
		{
			name: "custom configuration",
			config: ReconnectionConfig{
				MaxAttempts:   5,
				InitialDelay:  time.Second,
				BackoffMax:    time.Minute * 2,
				BackoffFactor: 2.0,
			},
		},
		{
			name: "infinite attempts configuration",
			config: ReconnectionConfig{
				MaxAttempts:   0, // Infinite
				InitialDelay:  500 * time.Millisecond,
				BackoffMax:    time.Minute,
				BackoffFactor: 1.8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm, _, _ := createMockConnectionManager(t, tt.config)

			if cm.maxReconnectAttempts != tt.config.MaxAttempts {
				t.Errorf("Expected MaxAttempts %d but got %d", tt.config.MaxAttempts, cm.maxReconnectAttempts)
			}
			if cm.reconnectDelay != tt.config.InitialDelay {
				t.Errorf("Expected InitialDelay %v but got %v", tt.config.InitialDelay, cm.reconnectDelay)
			}
			if cm.reconnectBackoffMax != tt.config.BackoffMax {
				t.Errorf("Expected BackoffMax %v but got %v", tt.config.BackoffMax, cm.reconnectBackoffMax)
			}
		})
	}
}

func TestConnectionManager_HealthMonitoring(t *testing.T) {
	t.Run("health monitoring setup", func(t *testing.T) {
		cm, mockConn, mockCh := createMockConnectionManager(t)

		// Verify notification channels are set up
		if cm.connCloseNotify == nil {
			t.Error("Expected connection close notify channel to be set up")
		}
		if cm.chCloseNotify == nil {
			t.Error("Expected channel close notify channel to be set up")
		}
		if cm.chCancelNotify == nil {
			t.Error("Expected channel cancel notify channel to be set up")
		}

		// Verify the channels have capacity
		if cap(cm.connCloseNotify) != 1 {
			t.Errorf("Expected connection close notify channel capacity 1 but got %d", cap(cm.connCloseNotify))
		}
		if cap(cm.chCloseNotify) != 1 {
			t.Errorf("Expected channel close notify channel capacity 1 but got %d", cap(cm.chCloseNotify))
		}
		if cap(cm.chCancelNotify) != 1 {
			t.Errorf("Expected channel cancel notify channel capacity 1 but got %d", cap(cm.chCancelNotify))
		}

		// Test that IsHealthy correctly evaluates the state
		if !cm.IsHealthy() {
			t.Error("Expected connection manager to be healthy initially")
		}

		// Test unhealthy states
		mockConn.Close()
		if cm.IsHealthy() {
			t.Error("Expected connection manager to be unhealthy when connection is closed")
		}

		// Reset connection, close channel
		mockConn = NewMockRMQConnection()
		cm.conn = mockConn
		mockCh.Close()
		if cm.IsHealthy() {
			t.Error("Expected connection manager to be unhealthy when channel is closed")
		}
	})
}

func TestConnectionManager_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent access to connection manager", func(t *testing.T) {
		cm, _, _ := createMockConnectionManager(t)

		// Test concurrent access to various methods
		var wg sync.WaitGroup
		iterations := 100

		// Concurrent GetConnection calls
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = cm.GetConnection()
			}()
		}

		// Concurrent GetChannel calls
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = cm.GetChannel()
			}()
		}

		// Concurrent IsHealthy calls
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cm.IsHealthy()
			}()
		}

		// Concurrent GetConnectionString calls
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cm.GetConnectionString()
			}()
		}

		wg.Wait()

		// If we get here without data races, the test passes
	})
}

func TestDefaultReconnectionConfig(t *testing.T) {
	expected := ReconnectionConfig{
		MaxAttempts:   0,               // Infinite attempts
		InitialDelay:  time.Second * 2, // Start with 2 second delay
		BackoffMax:    time.Minute * 5, // Maximum 5 minute delay
		BackoffFactor: 1.5,             // 1.5x exponential backoff
	}

	if DefaultReconnectionConfig.MaxAttempts != expected.MaxAttempts {
		t.Errorf("Expected MaxAttempts %d but got %d", expected.MaxAttempts, DefaultReconnectionConfig.MaxAttempts)
	}
	if DefaultReconnectionConfig.InitialDelay != expected.InitialDelay {
		t.Errorf("Expected InitialDelay %v but got %v", expected.InitialDelay, DefaultReconnectionConfig.InitialDelay)
	}
	if DefaultReconnectionConfig.BackoffMax != expected.BackoffMax {
		t.Errorf("Expected BackoffMax %v but got %v", expected.BackoffMax, DefaultReconnectionConfig.BackoffMax)
	}
	if DefaultReconnectionConfig.BackoffFactor != expected.BackoffFactor {
		t.Errorf("Expected BackoffFactor %f but got %f", expected.BackoffFactor, DefaultReconnectionConfig.BackoffFactor)
	}
}
