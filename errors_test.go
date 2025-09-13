// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"errors"
	"testing"
)

func TestBunMQError_Error(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "simple error message",
			message:  "connection failed",
			expected: "connection failed",
		},
		{
			name:     "empty error message",
			message:  "",
			expected: "",
		},
		{
			name:     "complex error message",
			message:  "failed to establish connection to rabbitmq://localhost:5672",
			expected: "failed to establish connection to rabbitmq://localhost:5672",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &BunMQError{Message: tt.message}
			if got := err.Error(); got != tt.expected {
				t.Errorf("BunMQError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewBunMQError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "create error with message",
			message:  "test error",
			expected: "test error",
		},
		{
			name:     "create error with empty message",
			message:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewBunMQError(tt.message)
			if err == nil {
				t.Fatal("NewBunMQError() returned nil")
			}
			if err.Message != tt.expected {
				t.Errorf("NewBunMQError().Message = %v, want %v", err.Message, tt.expected)
			}
			if err.Error() != tt.expected {
				t.Errorf("NewBunMQError().Error() = %v, want %v", err.Error(), tt.expected)
			}
		})
	}
}

func TestErrorFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(error) error
		input    error
		expected string
	}{
		{
			name:     "rabbitMQDialError wraps error",
			fn:       rabbitMQDialError,
			input:    errors.New("dial tcp: connection refused"),
			expected: "dial tcp: connection refused",
		},
		{
			name:     "getChannelError wraps error",
			fn:       getChannelError,
			input:    errors.New("channel creation failed"),
			expected: "channel creation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.input)
			if err == nil {
				t.Fatal("error function returned nil")
			}
			if err.Error() != tt.expected {
				t.Errorf("error function result = %v, want %v", err.Error(), tt.expected)
			}

			// Verify it's a BunMQError
			bunErr, ok := err.(*BunMQError)
			if !ok {
				t.Errorf("error function should return *BunMQError, got %T", err)
			}
			if bunErr.Message != tt.expected {
				t.Errorf("BunMQError.Message = %v, want %v", bunErr.Message, tt.expected)
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "NullableChannelError",
			err:      NullableChannelError,
			expected: "channel cant be null",
		},
		{
			name:     "ConsumerAlreadyRegisteredForTheMessageError",
			err:      ConsumerAlreadyRegisteredForTheMessageError,
			expected: "consumer already registered for the message",
		},
		{
			name:     "NotFoundQueueDefinitionError",
			err:      NotFoundQueueDefinitionError,
			expected: "not found queue definition",
		},
		{
			name:     "InvalidDispatchParamsError",
			err:      InvalidDispatchParamsError,
			expected: "register dispatch with invalid parameters",
		},
		{
			name:     "QueueDefinitionNotFoundError",
			err:      QueueDefinitionNotFoundError,
			expected: "any queue definition was founded to the given queue",
		},
		{
			name:     "ReceivedMessageWithUnformattedHeaderError",
			err:      ReceivedMessageWithUnformattedHeaderError,
			expected: "received message with unformatted headers",
		},
		{
			name:     "RetryableError",
			err:      RetryableError,
			expected: "error to process this message, retry latter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("predefined error is nil")
			}
			if tt.err.Error() != tt.expected {
				t.Errorf("error message = %v, want %v", tt.err.Error(), tt.expected)
			}

			// Verify it's a BunMQError
			bunErr, ok := tt.err.(*BunMQError)
			if !ok {
				t.Errorf("predefined error should be *BunMQError, got %T", tt.err)
			}
			if bunErr.Message != tt.expected {
				t.Errorf("BunMQError.Message = %v, want %v", bunErr.Message, tt.expected)
			}
		})
	}
}

func TestBunMQError_Interface(t *testing.T) {
	// Verify BunMQError implements error interface
	var err error = &BunMQError{Message: "test"}
	if err.Error() != "test" {
		t.Errorf("BunMQError does not properly implement error interface")
	}
}