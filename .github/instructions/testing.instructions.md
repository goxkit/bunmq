---
applyTo: "**/*_test.go"
---

# Testing Instructions

## Rule: Test File Naming and Location

**Description**: Test files MUST be named with `*_test.go` suffix and placed alongside source files.

**When it applies**: When creating test files.

**Copilot MUST**:
- Name test files with `*_test.go` suffix
- Place test files in the same directory as the code being tested
- Use `package bunmq` for internal tests (access to private members)
- Use descriptive test function names: `Test<FunctionName>_<Scenario>`

**Copilot MUST NOT**:
- Create test files without `_test.go` suffix
- Put tests in separate directories
- Use generic test function names

**Example - Correct Test File Organization**:

Reference: bunmq library structure

```
bunmq/
├── publisher.go           # Implementation
├── publisher_test.go      # ✅ Tests alongside source file
├── dispatcher.go          # Implementation
├── dispatcher_test.go     # ✅ Tests alongside source file
├── connection_manager.go  # Implementation
└── connection_manager_test.go  # ✅ Tests alongside source file
```

---

## Rule: Table-Driven Tests

**Description**: Use table-driven tests for multiple test cases.

**When it applies**: When testing functions with multiple scenarios.

**Copilot MUST**:
- Use table-driven tests for similar test cases
- Structure test tables with clear field names
- Use subtests with `t.Run()` for each test case
- Include setup, execution, and assertion in each test case

**Copilot MUST NOT**:
- Duplicate test code for similar cases
- Create separate test functions for similar scenarios
- Mix different test patterns in one file

**Example - Correct Table-Driven Test**:

Reference: `@publisher_test.go`

```go
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
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            manager := NewMockConnectionManager()
            channel := NewMockAMQPChannel()
            if tt.channelErr != nil {
                manager.SetGetChannelError(tt.channelErr)
            } else {
                manager.SetChannel(channel)
                channel.SetPublishError(tt.publishErr)
            }

            publisher := NewPublisher("test-app", manager)

            // Execute
            err := publisher.Publish(context.Background(), tt.exchange, tt.routingKey, tt.msg)

            // Assert
            if (err != nil) != tt.expectError {
                t.Errorf("Publish() error = %v, expectError %v", err, tt.expectError)
            }
        })
    }
}
```

Reference: `@dispatcher_test.go`

```go
func TestDispatcher_RegisterByType(t *testing.T) {
    tests := []struct {
        name          string
        queue         string
        msg           interface{}
        handler       ConsumerHandler
        existingQueue bool
        expectError   bool
        expectedError string
    }{
        {
            name:          "successful registration",
            queue:         "test-queue",
            msg:           DispatcherTestMessage{},
            handler:       func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil },
            existingQueue: true,
            expectError:   false,
        },
        {
            name:          "nil message",
            queue:         "test-queue",
            msg:           nil,
            handler:       func(ctx context.Context, msg any, metadata *DeliveryMetadata) error { return nil },
            existingQueue: true,
            expectError:   true,
            expectedError: "register dispatch with invalid parameters",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... test implementation
        })
    }
}
```

---

## Rule: Mock Usage with In-Package Mocks

**Description**: Use the mock implementations defined in `mocks.go` for testing.

**When it applies**: When testing code that depends on interfaces (connections, channels).

**Copilot MUST**:
- Use mock types from `mocks.go`: `MockAMQPChannel`, `MockConnectionManager`, `MockRMQConnection`
- Configure mock behavior using setter methods (e.g., `SetPublishError`)
- Verify mock interactions using helper methods (e.g., `GetPublishedMessages`)
- Use `NewMock*` constructors to create mocks with default values

**Copilot MUST NOT**:
- Create duplicate mock implementations
- Test implementation details through mocks
- Use real connections in unit tests

**Example - Correct Mock Usage**:

Reference: `@mocks.go`

```go
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
    qosError             error
    closed               bool
    closeError           error
    // ...
}

func NewMockAMQPChannel() *MockAMQPChannel {
    deliveryChannel := make(chan amqp.Delivery)
    close(deliveryChannel)

    return &MockAMQPChannel{
        closed:               false,
        notifyCloseChannels:  make([]chan *amqp.Error, 0),
        notifyCancelChannels: make([]chan string, 0),
        queueDeclareQueue:    amqp.Queue{Name: "test-queue", Messages: 0, Consumers: 0},
        consumeChannel:       deliveryChannel,
        publishedMessages:    make([]PublishedMessage, 0),
    }
}
```

**Test Usage**:

Reference: `@publisher_test.go`

```go
func TestPublisher_Publish_VerifyMessageContent(t *testing.T) {
    // ✅ Create mock instances
    manager := NewMockConnectionManager()
    channel := NewMockAMQPChannel()
    manager.SetChannel(channel)

    publisher := NewPublisher("test-app", manager)
    testMsg := TestMessage{ID: "123", Content: "hello"}

    err := publisher.Publish(context.Background(), "exchange", "key", testMsg)
    if err != nil {
        t.Fatalf("Publish() unexpected error: %v", err)
    }

    // ✅ Verify published message using mock helper
    messages := channel.GetPublishedMessages()
    if len(messages) != 1 {
        t.Fatalf("expected 1 published message, got %d", len(messages))
    }

    // ✅ Verify message content
    if messages[0].Exchange != "exchange" {
        t.Errorf("expected exchange 'exchange', got '%s'", messages[0].Exchange)
    }
}
```

---

## Rule: Test Coverage Requirements

**Description**: Maintain 90%+ test coverage for library code.

**When it applies**: When generating test code.

**Copilot MUST**:
- Write tests for all public functions and methods
- Write tests for error handling paths
- Write tests for edge cases (nil inputs, empty values)
- Write tests for concurrent access scenarios
- Run `make test-coverage-threshold` to verify coverage

**Copilot MUST NOT**:
- Test only happy paths
- Ignore error cases
- Skip testing edge cases
- Commit code with coverage below 90%

**Example - Coverage Commands**:

Reference: `@Makefile`

```makefile
# Run tests with coverage
make test-coverage

# Check coverage threshold (90%)
make test-coverage-threshold
```

---

## Rule: Test Message Types

**Description**: Define test-specific message types in test files.

**When it applies**: When testing message handling (publisher, dispatcher).

**Copilot MUST**:
- Define test message types at the top of test files
- Use simple, clear struct definitions
- Include JSON tags for serialization tests

**Copilot MUST NOT**:
- Use production message types in tests (unless testing specific compatibility)
- Create overly complex test message types

**Example - Correct Test Message Types**:

Reference: `@publisher_test.go`

```go
// Test message types
type TestMessage struct {
    ID      string `json:"id"`
    Content string `json:"content"`
}

type TestPointerMessage struct {
    Value int `json:"value"`
}
```

Reference: `@dispatcher_test.go`

```go
// Test message types for dispatcher tests
type DispatcherTestMessage struct {
    ID      string `json:"id"`
    Content string `json:"content"`
}

type AnotherTestMessage struct {
    Name  string `json:"name"`
    Value int    `json:"value"`
}
```

---

## Rule: Test Setup and Cleanup

**Description**: Use proper setup and cleanup patterns in tests.

**When it applies**: When tests require resources or complex setup.

**Copilot MUST**:
- Use `t.Cleanup()` for cleanup operations
- Use helper functions for common setup
- Use `defer` for resource cleanup within tests
- Cancel contexts properly in tests

**Copilot MUST NOT**:
- Leave resources uncleaned
- Create global test state
- Share state between tests without proper isolation

**Example - Correct Test Setup**:

```go
func TestDispatcher_Consume(t *testing.T) {
    // Setup
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    t.Cleanup(func() {
        cancel()  // ✅ Ensure context is cancelled
    })

    manager := NewMockConnectionManager()
    channel := NewMockAMQPChannel()
    manager.SetChannel(channel)

    // Setup delivery channel
    deliveryCh := make(chan amqp.Delivery)
    channel.SetConsumeChannel(deliveryCh)
    t.Cleanup(func() {
        close(deliveryCh)  // ✅ Clean up channel
    })

    // ... test execution
}
```

---

## Rule: Concurrent Test Safety

**Description**: Ensure tests are safe for parallel execution.

**When it applies**: When testing concurrent code or using `t.Parallel()`.

**Copilot MUST**:
- Use `t.Parallel()` for independent tests
- Use local variables (not shared state) in parallel tests
- Use `sync.WaitGroup` for coordinating concurrent operations in tests
- Capture loop variables in closures

**Copilot MUST NOT**:
- Share mutable state between parallel tests
- Use global variables in tests
- Create race conditions in tests

**Example - Correct Concurrent Test**:

Reference: `@dispatcher_test.go`

```go
func TestDispatcher_ConcurrentConsumption(t *testing.T) {
    manager := NewMockConnectionManager()
    channel := NewMockAMQPChannel()
    manager.SetChannel(channel)

    dispatcher := NewDispatcher(manager, []*QueueDefinition{
        NewQueue("test-queue"),
    })

    var wg sync.WaitGroup
    processedCount := 0
    var mu sync.Mutex

    handler := func(ctx context.Context, msg any, metadata *DeliveryMetadata) error {
        mu.Lock()
        processedCount++
        mu.Unlock()
        return nil
    }

    err := dispatcher.RegisterByType("test-queue", DispatcherTestMessage{}, handler)
    if err != nil {
        t.Fatalf("RegisterByType() error: %v", err)
    }

    // ... test concurrent message processing
}
```

---

## Rule: Error Path Testing

**Description**: Thoroughly test all error paths and edge cases.

**When it applies**: When testing functions that can return errors.

**Copilot MUST**:
- Test all error return paths
- Test nil input handling
- Test empty/zero value handling
- Test timeout scenarios
- Test connection/channel failure scenarios

**Copilot MUST NOT**:
- Ignore error paths
- Assume happy path only
- Skip testing edge cases

**Example - Error Path Testing**:

Reference: `@publisher_test.go`

```go
func TestPublisher_Publish_ErrorCases(t *testing.T) {
    tests := []struct {
        name        string
        setup       func(*MockConnectionManager, *MockAMQPChannel)
        exchange    string
        expectError bool
    }{
        {
            name: "channel error",
            setup: func(m *MockConnectionManager, ch *MockAMQPChannel) {
                m.SetGetChannelError(errors.New("channel error"))
            },
            exchange:    "test-exchange",
            expectError: true,
        },
        {
            name: "publish error",
            setup: func(m *MockConnectionManager, ch *MockAMQPChannel) {
                ch.SetPublishError(errors.New("publish error"))
                m.SetChannel(ch)
            },
            exchange:    "test-exchange",
            expectError: true,
        },
        {
            name: "empty exchange",
            setup: func(m *MockConnectionManager, ch *MockAMQPChannel) {
                m.SetChannel(ch)
            },
            exchange:    "",
            expectError: true,  // ✅ Validate input error
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := NewMockConnectionManager()
            channel := NewMockAMQPChannel()
            if tt.setup != nil {
                tt.setup(manager, channel)
            }

            publisher := NewPublisher("test-app", manager)
            err := publisher.Publish(context.Background(), tt.exchange, "key", TestMessage{})

            if (err != nil) != tt.expectError {
                t.Errorf("Publish() error = %v, expectError %v", err, tt.expectError)
            }
        })
    }
}
```

---

## Rule: Test Assertions

**Description**: Use clear assertions with the standard library or consistent patterns.

**When it applies**: When writing test assertions.

**Copilot MUST**:
- Use standard library testing (`t.Error`, `t.Errorf`, `t.Fatal`, `t.Fatalf`)
- Use `t.Helper()` for custom assertion helpers
- Provide descriptive error messages
- Use `errors.Is()` for error comparisons

**Copilot MUST NOT**:
- Use generic assertions without messages
- Panic in assertions (use `t.Fatal` instead)
- Ignore assertion failures

**Example - Correct Assertions**:

```go
func TestPublisher_Publish(t *testing.T) {
    manager := NewMockConnectionManager()
    channel := NewMockAMQPChannel()
    manager.SetChannel(channel)

    publisher := NewPublisher("test-app", manager)
    err := publisher.Publish(context.Background(), "exchange", "key", TestMessage{ID: "1"})

    // ✅ Clear assertion with context
    if err != nil {
        t.Fatalf("Publish() unexpected error: %v", err)
    }

    // ✅ Verify behavior
    messages := channel.GetPublishedMessages()
    if len(messages) == 0 {
        t.Fatal("expected at least one published message")
    }

    // ✅ Compare with descriptive message
    if messages[0].Exchange != "exchange" {
        t.Errorf("expected exchange 'exchange', got '%s'", messages[0].Exchange)
    }
}

// ✅ Custom assertion helper
func assertNoError(t *testing.T, err error, msg string) {
    t.Helper()
    if err != nil {
        t.Fatalf("%s: unexpected error: %v", msg, err)
    }
}
```

---

## Testing Checklist

When generating tests for bunmq, ensure:

- ✅ Test file named `*_test.go` alongside source file
- ✅ Package is `bunmq` (internal tests)
- ✅ Table-driven tests for multiple scenarios
- ✅ Mocks from `mocks.go` used correctly
- ✅ Test message types defined at top of file
- ✅ Both success and error paths tested
- ✅ Edge cases tested (nil, empty, zero values)
- ✅ Concurrent scenarios tested where relevant
- ✅ Resources cleaned up with `t.Cleanup()` or `defer`
- ✅ Coverage maintained at 90%+ (`make test-coverage-threshold`)

---

## Anti-Patterns (FORBIDDEN)

**Copilot MUST NEVER generate**:

```go
// ❌ Test without subtests for multiple cases
func TestPublisher(t *testing.T) {
    // Multiple unrelated assertions
}

// ❌ Creating separate mock implementations
type myMockChannel struct {}  // ❌ Use MockAMQPChannel from mocks.go

// ❌ Ignoring errors in test setup
manager, _ := NewConnectionManager(ctx, "amqp://", "app")  // ❌ Check errors

// ❌ Shared mutable state
var globalMock *MockAMQPChannel  // ❌ Create in each test

// ❌ Missing cleanup
func TestSomething(t *testing.T) {
    ch := make(chan amqp.Delivery)
    // ❌ Channel never closed
}
```

---

## Positive Testing Patterns (REQUIRED)

**Copilot MUST generate**:

```go
// ✅ Table-driven test with clear structure
func TestFunction_Scenario(t *testing.T) {
    tests := []struct{
        name        string
        setup       func(*MockConnectionManager)
        input       string
        expectError bool
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            manager := NewMockConnectionManager()
            if tt.setup != nil {
                tt.setup(manager)
            }

            // Act
            result, err := Function(tt.input)

            // Assert
            if (err != nil) != tt.expectError {
                t.Errorf("error = %v, expectError %v", err, tt.expectError)
            }
        })
    }
}

// ✅ Proper cleanup
func TestWithCleanup(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    t.Cleanup(cancel)

    ch := make(chan amqp.Delivery)
    t.Cleanup(func() { close(ch) })

    // ... test
}
```
