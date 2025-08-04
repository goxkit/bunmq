// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

// RabbitMQError represents a custom error type for RabbitMQ-related operations.
// It encapsulates an error message describing the specific error condition.
type BunMQError struct {
	Message string
}

// NewBunMQError creates a new BunMQError instance with the provided message.
// Returns the error as an error interface.
func NewBunMQError(msg string) *BunMQError {
	return &BunMQError{Message: msg}
}

// Error implements the error interface and returns the error message.
func (e *BunMQError) Error() string {
	return e.Message
}

var (
	// rabbitMQDialError is a function that wraps a connection error into a BunMQError.
	rabbitMQDialError = func(err error) error { return NewBunMQError(err.Error()) }

	// getChannelError is a function that wraps a channel creation error into a BunMQError.
	getChannelError = func(err error) error { return NewBunMQError(err.Error()) }

	// NullableChannelError is returned when a channel operation is attempted on a nil channel.
	NullableChannelError = NewBunMQError("channel cant be null")

	// ConsumerAlreadyRegisteredError is returned when a consumer is already registered for a specific queue.
	ConsumerAlreadyRegisteredForTheMessageError = NewBunMQError("consumer already registered for the message")

	// NotFoundQueueDefinitionError is returned when a queue definition cannot be found.
	NotFoundQueueDefinitionError = NewBunMQError("not found queue definition")

	// InvalidDispatchParamsError is returned when invalid parameters are provided to a dispatch operation.
	InvalidDispatchParamsError = NewBunMQError("register dispatch with invalid parameters")

	// QueueDefinitionNotFoundError is returned when no queue definition is found for a specified queue.
	QueueDefinitionNotFoundError = NewBunMQError("any queue definition was founded to the given queue")

	// ReceivedMessageWithUnformattedHeaderError is returned when a message has incorrectly formatted headers.
	ReceivedMessageWithUnformattedHeaderError = NewBunMQError("received message with unformatted headers")

	// RetryableError indicates that a message processing failed but can be retried later.
	RetryableError = NewBunMQError("error to process this message, retry latter")
)
