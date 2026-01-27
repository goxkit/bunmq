---
applyTo: "**/*.go"
---

# Architecture Instructions

## Rule: Interface-Based Design

**Description**: All major components MUST be defined as interfaces for testability and flexibility.

**When it applies**: When creating new components or modifying existing ones.

**Copilot MUST**:
- Define public behavior as interfaces
- Keep interfaces small and focused (Interface Segregation Principle)
- Place interface definitions with the type that depends on them (consumer side)
- Document interface methods with clear descriptions

**Copilot MUST NOT**:
- Create large monolithic interfaces
- Put interface definitions in implementation files
- Create interfaces that expose implementation details
- Skip interface documentation

**Example - Correct Interface Design**:

Reference: `@connection.go`

```go
// ✅ CORRECT: Small, focused interface with documentation
type RMQConnection interface {
    // Channel creates a new channel on the connection.
    Channel() (*amqp.Channel, error)

    // ConnectionState returns the TLS connection state if TLS is enabled.
    ConnectionState() tls.ConnectionState

    // IsClosed checks if the connection is closed.
    IsClosed() bool

    // Close gracefully closes the connection.
    Close() error

    // NotifyClose returns a channel for connection close notifications.
    NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
}
```

Reference: `@publisher.go`

```go
// ✅ CORRECT: Clear interface for publishing
type Publisher interface {
    Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error
    PublishDeadline(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error
    PublishQueue(ctx context.Context, queue string, msg any, options ...*Option) error
    PublishQueueDeadline(ctx context.Context, queue string, msg any, options ...*Option) error
}
```

**Example - Incorrect Interface Design**:

```go
// ❌ WRONG: God interface with too many responsibilities
type MessageBroker interface {
    Connect() error
    Disconnect() error
    CreateChannel() error
    DeclareExchange() error
    DeclareQueue() error
    BindQueue() error
    Publish() error
    Consume() error
    Acknowledge() error
    Reject() error
    // ... too many methods
}
```

---

## Rule: Builder Pattern for Complex Configuration

**Description**: Use builder pattern for objects with many optional configuration options.

**When it applies**: When creating types that need flexible configuration.

**Copilot MUST**:
- Use method chaining for optional settings
- Provide sensible defaults
- Return the builder/self for chaining
- Validate configuration at build time

**Copilot MUST NOT**:
- Create constructors with many parameters
- Require all configuration upfront
- Skip validation of configuration

**Example - Correct Builder Pattern**:

Reference: `@queue.go`

```go
// ✅ CORRECT: Builder pattern with method chaining
func NewQueue(name string) *QueueDefinition {
    return &QueueDefinition{
        name:    name,
        durable: false,
    }
}

func (q *QueueDefinition) Durable(durable bool) *QueueDefinition {
    q.durable = durable
    return q
}

func (q *QueueDefinition) WithMaxLength(maxLength int) *QueueDefinition {
    q.maxLength = maxLength
    return q
}

func (q *QueueDefinition) WithRetry(delay time.Duration, retries int) *QueueDefinition {
    q.retryDelay = delay
    q.maxRetries = retries
    return q
}

func (q *QueueDefinition) WithDLQ() *QueueDefinition {
    q.useDLQ = true
    return q
}
```

**Usage**:

```go
// ✅ CORRECT: Fluent configuration
queueDef := bunmq.NewQueue("orders").
    Durable(true).
    WithMaxLength(100_000).
    WithRetry(time.Second*10, 3).
    WithDLQ()
```

---

## Rule: Dependency Injection via Constructor

**Description**: All dependencies MUST be injected via constructor parameters.

**When it applies**: When creating new structs that depend on external resources.

**Copilot MUST**:
- Accept dependencies as constructor parameters
- Use interfaces for all dependencies (Dependency Inversion)
- Return interfaces, not concrete types
- Document constructor parameters

**Copilot MUST NOT**:
- Create dependencies inside methods
- Use global static instances
- Hardcode concrete implementations
- Hide dependency creation inside constructors

**Example - Correct Dependency Injection**:

Reference: `@publisher.go`

```go
// ✅ CORRECT: Dependencies injected via constructor
type publisher struct {
    appName string
    manager ConnectionManager  // Interface, not concrete type
}

func NewPublisher(appName string, manager ConnectionManager) Publisher {
    return &publisher{appName, manager}
}
```

Reference: `@dispatcher.go`

```go
// ✅ CORRECT: Multiple dependencies injected
type dispatcher struct {
    manager             ConnectionManager
    queueDefinitions    map[string]*QueueDefinition
    consumersDefinition map[string][]*ConsumerDefinition
    tracer              trace.Tracer
    signalCh            chan os.Signal
    mu                  sync.RWMutex
}

func NewDispatcher(manager ConnectionManager, definitions []*QueueDefinition) Dispatcher {
    // ...
}
```

**Example - Incorrect Pattern**:

```go
// ❌ WRONG: Creating dependencies internally
type badPublisher struct{}

func NewBadPublisher() *badPublisher {
    return &badPublisher{}
}

func (p *badPublisher) Publish(msg any) error {
    // ❌ Creating connection inside method
    conn, _ := amqp.Dial("amqp://localhost")
    ch, _ := conn.Channel()
    // ...
}
```

---

## Rule: Single Responsibility Principle

**Description**: Each struct, function, and file MUST have ONE clear responsibility.

**When it applies**: When generating any new code.

**Copilot MUST**:
- Keep functions under 30 lines (excluding comments)
- Ensure each struct has a single purpose
- Split files by concept (one main type per file)
- Extract complex logic to helper functions

**Copilot MUST NOT**:
- Create god objects or god functions
- Mix unrelated responsibilities
- Create functions over 50 lines
- Allow deep nesting (> 3 levels)

**Example - Correct Single Responsibility**:

```
bunmq/
├── connection.go           # Only connection interface
├── connection_manager.go   # Only connection management
├── publisher.go            # Only publishing
├── dispatcher.go           # Only dispatching/consuming
├── topology.go             # Only topology management
├── queue.go                # Only queue definitions
├── exchange.go             # Only exchange definitions
└── tracing.go              # Only tracing utilities
```

Reference: `@tracing.go`

```go
// ✅ CORRECT: Single responsibility - trace context propagation
type AMQPHeader amqp.Table

func (h AMQPHeader) Set(key, val string) {
    key = strings.ToLower(key)
    h[key] = val
}

func (h AMQPHeader) Get(key string) string {
    key = strings.ToLower(key)
    value, ok := h[key]
    if !ok {
        return ""
    }
    toString, ok := value.(string)
    if !ok {
        return ""
    }
    return toString
}

func (h AMQPHeader) Keys() []string {
    keys := make([]string, 0, len(h))
    for k := range h {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    return keys
}
```

---

## Rule: Context Propagation

**Description**: Always propagate `context.Context` through all function calls.

**When it applies**: When creating functions that perform I/O or can be cancelled.

**Copilot MUST**:
- Accept `context.Context` as the first parameter
- Propagate context through all layers
- Use context for tracing, logging, and cancellation
- Create spans using context

**Copilot MUST NOT**:
- Create new contexts without reason
- Store context in structs
- Ignore context cancellation
- Pass `nil` context

**Example - Correct Context Propagation**:

Reference: `@publisher.go`

```go
// ✅ CORRECT: Context as first parameter, propagated
func (p *publisher) Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error {
    // Use context for logging
    if exchange == "" {
        logrus.WithContext(ctx).Error("exchange cannot be empty")
        return fmt.Errorf("exchange cannot be empty")
    }
    // Propagate context
    return p.publish(ctx, exchange, routingKey, msg, options...)
}
```

Reference: `@dispatcher.go`

```go
// ✅ CORRECT: Context propagated to handler
func (d *dispatcher) handleDelivery(ctx context.Context, delivery amqp.Delivery, def *ConsumerDefinition) error {
    ctx, span := d.tracer.Start(ctx, "dispatcher.handleDelivery")
    defer span.End()

    // Context propagated to user handler
    return def.handler(ctx, msg, metadata)
}
```

---

## Rule: Error Handling and Domain Errors

**Description**: Use domain-specific error types and handle errors explicitly.

**When it applies**: When handling any error condition.

**Copilot MUST**:
- Define domain-specific error types
- Return errors, never panic in library code
- Wrap errors with context when needed
- Log errors before returning them

**Copilot MUST NOT**:
- Ignore errors with `_`
- Use generic error strings
- Lose error context
- Panic for recoverable errors

**Example - Correct Error Handling**:

Reference: `@errors.go`

```go
// ✅ CORRECT: Domain-specific error variables
var (
    ErrQueueNotDefined       = errors.New("queue not defined in topology")
    ErrHandlerAlreadyExists  = errors.New("handler already exists for this queue")
    ErrConnectionClosed      = errors.New("connection is closed")
    ErrChannelClosed         = errors.New("channel is closed")
)
```

Reference: `@publisher.go`

```go
// ✅ CORRECT: Input validation with clear error
func (p *publisher) Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error {
    if exchange == "" {
        logrus.WithContext(ctx).Error("exchange cannot be empty")
        return fmt.Errorf("exchange cannot be empty")
    }
    return p.publish(ctx, exchange, routingKey, msg, options...)
}
```

---

## Rule: Thread Safety with Synchronization

**Description**: Ensure concurrent-safe code with proper synchronization primitives.

**When it applies**: When code may be accessed from multiple goroutines.

**Copilot MUST**:
- Use `sync.Mutex` or `sync.RWMutex` for shared state
- Protect all concurrent access to shared maps and slices
- Use channels for communication between goroutines
- Document thread-safety guarantees

**Copilot MUST NOT**:
- Access shared state without synchronization
- Create race conditions
- Hold locks longer than necessary
- Use global mutable state

**Example - Correct Thread Safety**:

Reference: `@dispatcher.go`

```go
// ✅ CORRECT: Mutex for thread-safe access
type dispatcher struct {
    consumersDefinition map[string][]*ConsumerDefinition
    mu                  sync.RWMutex  // Protects consumersDefinition
}

func (d *dispatcher) RegisterByType(queue string, typE any, handler ConsumerHandler) error {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Safe to modify map
    d.consumersDefinition[queue] = append(d.consumersDefinition[queue], def)
    return nil
}
```

---

## Rule: Graceful Shutdown

**Description**: Handle shutdown signals properly for clean resource cleanup.

**When it applies**: When code manages long-running processes or connections.

**Copilot MUST**:
- Listen for SIGTERM and SIGINT signals
- Provide graceful shutdown mechanisms
- Clean up resources (connections, channels)
- Use context cancellation for propagation

**Copilot MUST NOT**:
- Ignore shutdown signals
- Leave resources open
- Block shutdown indefinitely
- Skip cleanup on exit

**Example - Correct Graceful Shutdown**:

Reference: `@dispatcher.go`

```go
// ✅ CORRECT: Signal handling for graceful shutdown
func (d *dispatcher) ConsumeBlocking() {
    signalCh := make(chan os.Signal, 1)
    signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

    // Start consumers...

    // Block until signal
    <-signalCh

    // Graceful shutdown
    logrus.Info("shutting down dispatcher...")
    // Cleanup resources
}
```

---

## Architecture Checklist

When generating code, ensure:

- ✅ Interfaces defined for public behavior
- ✅ Builder pattern used for complex configuration
- ✅ Dependencies injected via constructors
- ✅ Single responsibility maintained
- ✅ Context propagated as first parameter
- ✅ Domain-specific errors used
- ✅ Thread safety with proper synchronization
- ✅ Graceful shutdown implemented
