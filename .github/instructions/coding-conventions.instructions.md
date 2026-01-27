---
applyTo: "**/*.go"
---

# Coding Conventions Instructions

## Rule: Naming Conventions

**Description**: All identifiers MUST follow Go naming conventions strictly.

**When it applies**: When naming any identifier (types, functions, variables, constants).

**Copilot MUST**:
- Use PascalCase for exported names (types, functions, constants)
- Use camelCase for unexported names
- Use descriptive, intention-revealing names
- Follow domain terminology consistently
- Use conventional short names: `ctx`, `err`, `msg`, `ch`

**Copilot MUST NOT**:
- Use single-letter names (except `i`, `j`, `k` for iterators)
- Use abbreviations unless universally understood
- Use generic names (`data`, `info`, `temp`, `obj`)
- Mix naming conventions

**Example - Correct Naming**:

Reference: `@connection.go`

```go
// ✅ CORRECT: Exported interface with PascalCase and documentation
type RMQConnection interface {
    Channel() (*amqp.Channel, error)
    IsClosed() bool
    Close() error
    NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
}
```

Reference: `@dispatcher.go`

```go
// ✅ CORRECT: Unexported struct with camelCase
type dispatcher struct {
    manager             ConnectionManager
    queueDefinitions    map[string]*QueueDefinition
    consumersDefinition map[string][]*ConsumerDefinition
    tracer              trace.Tracer
    signalCh            chan os.Signal
    mu                  sync.RWMutex
}

// ✅ CORRECT: Constructor with New prefix
func NewDispatcher(manager ConnectionManager, definitions []*QueueDefinition) Dispatcher {
    return &dispatcher{...}
}

// ✅ CORRECT: Method with clear verb action
func (d *dispatcher) RegisterByType(queue string, typE any, handler ConsumerHandler) error {
    // ...
}
```

**Example - Incorrect Naming**:

```go
// ❌ WRONG: Single-letter variable for important data
func Process(m *Message) error {
    // ...
}

// ❌ WRONG: Generic name
func Handle(data interface{}) error {
    // ...
}

// ❌ WRONG: Abbreviation without context
func PubMsg(msg *Message) error {
    // ...
}
```

---

## Rule: Function Size and Complexity

**Description**: Functions MUST be small, focused, and under 30 lines (excluding comments).

**When it applies**: When generating any function.

**Copilot MUST**:
- Keep functions under 30 lines (excluding comments and blank lines)
- Extract complex logic to helper functions
- Use early returns (guard clauses) to reduce nesting
- Ensure single responsibility per function

**Copilot MUST NOT**:
- Create functions over 50 lines
- Allow deep nesting (> 3 levels)
- Mix multiple concerns in one function
- Create god functions

**Example - Correct Function Size**:

Reference: `@publisher.go`

```go
// ✅ CORRECT: Small, focused function with early returns
func (p *publisher) Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error {
    if exchange == "" {
        logrus.WithContext(ctx).Error("exchange cannot be empty")
        return fmt.Errorf("exchange cannot be empty")
    }
    return p.publish(ctx, exchange, routingKey, msg, options...)
}
```

Reference: `@tracing.go`

```go
// ✅ CORRECT: Focused function with single responsibility
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
```

---

## Rule: Code Formatting

**Description**: Code MUST be formatted according to `gofmt` and `goimports` standards.

**When it applies**: When generating any Go code.

**Copilot MUST**:
- Format code according to `gofmt` standards
- Organize imports with `goimports` (stdlib, third-party)
- Use tabs for indentation
- Keep lines under 120 characters when possible

**Copilot MUST NOT**:
- Use non-standard formatting
- Mix tabs and spaces
- Create lines exceeding 120 characters unnecessarily

**Example - Correct Import Organization**:

Reference: `@dispatcher.go`

```go
import (
    // Standard library
    "context"
    "encoding/json"
    "errors"
    "os"
    "os/signal"
    "reflect"
    "sync"
    "syscall"
    "time"

    // Third-party dependencies
    amqp "github.com/rabbitmq/amqp091-go"
    "github.com/sirupsen/logrus"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)
```

---

## Rule: Comments and Documentation

**Description**: All exported symbols MUST have documentation comments.

**When it applies**: When creating exported types, functions, or packages.

**Copilot MUST**:
- Add package comments explaining package purpose
- Document all exported functions with clear descriptions
- Document all exported types and interfaces
- Start comments with the symbol name
- Explain "why" in comments, not "what"

**Copilot MUST NOT**:
- Skip documentation for exported symbols
- Add obvious comments that restate code
- Document implementation details unnecessarily

**Example - Correct Documentation**:

Reference: `@connection.go`

```go
// RMQConnection defines the interface for a RabbitMQ connection.
// It abstracts the underlying AMQP connection and provides methods
// for creating channels, retrieving connection state, and closing the connection.
type RMQConnection interface {
    // Channel creates a new channel on the connection.
    // Returns a pointer to an AMQP channel and any error encountered.
    Channel() (*amqp.Channel, error)

    // IsClosed checks if the connection is closed.
    IsClosed() bool

    // Close gracefully closes the connection and all its channels.
    Close() error

    // NotifyClose returns a channel that receives notifications when the connection is closed.
    NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
}
```

Reference: `@publisher.go`

```go
// Publisher defines an interface for publishing messages to a messaging system.
// It provides methods for sending messages with optional metadata such as
// destination, source, and routing keys.
type Publisher interface {
    // Publish sends a message to the specified exchange.
    Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error

    // PublishDeadline sends a message with a deadline.
    PublishDeadline(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error
}
```

---

## Rule: Magic Numbers and Constants

**Description**: NEVER use magic numbers. Always use named constants.

**When it applies**: When using numeric literals or string literals.

**Copilot MUST**:
- Define constants for all numeric literals (except 0, 1, -1)
- Use descriptive constant names
- Group related constants together
- Use constants for configuration values

**Copilot MUST NOT**:
- Use magic numbers in code
- Use string literals repeatedly for the same value
- Hardcode values that should be configurable

**Example - Correct Constant Usage**:

Reference: `@publisher.go`

```go
const (
    JSONContentType = "application/json"
)
```

Reference: `@dispatcher.go`

```go
const (
    ConsumerDefinitionByType ConsumerDefinitionType = iota + 1
    ConsumerDefinitionByExchange
    ConsumerDefinitionByRoutingKey
    ConsumerDefinitionByExchangeRoutingKey
)
```

**Example - Incorrect Usage**:

```go
// ❌ WRONG: Magic number
time.Sleep(5 * time.Second)  // What is 5? Why 5?

// ✅ CORRECT: Named constant
const reconnectDelay = 5 * time.Second
time.Sleep(reconnectDelay)
```

---

## Rule: Error Messages

**Description**: Error messages MUST be descriptive and actionable.

**When it applies**: When creating error messages or log messages.

**Copilot MUST**:
- Explain what went wrong
- Include relevant context
- Use consistent format
- Start error strings with lowercase (per Go convention)

**Copilot MUST NOT**:
- Use generic messages ("error occurred")
- Include sensitive information
- Use punctuation at the end of error messages

**Example - Correct Error Messages**:

Reference: `@errors.go`

```go
// ✅ CORRECT: Descriptive error variables
var (
    ErrQueueNotDefined       = errors.New("queue not defined in topology")
    ErrHandlerAlreadyExists  = errors.New("handler already exists for this queue")
    ErrConnectionClosed      = errors.New("connection is closed")
    ErrChannelClosed         = errors.New("channel is closed")
)
```

Reference: `@publisher.go`

```go
// ✅ CORRECT: Clear error with context
if exchange == "" {
    logrus.WithContext(ctx).Error("exchange cannot be empty")
    return fmt.Errorf("exchange cannot be empty")
}
```

---

## Rule: Struct Field Organization

**Description**: Struct fields MUST be organized logically.

**When it applies**: When creating structs.

**Copilot MUST**:
- Group related fields together
- Add field tags when needed (json, etc.)
- Align struct tags for readability
- Order fields by importance/usage

**Example - Correct Struct Organization**:

Reference: `@dispatcher.go`

```go
// ✅ CORRECT: Logical field grouping
type ConsumerDefinition struct {
    typ        ConsumerDefinitionType  // Type information
    queue      string                  // Queue binding
    exchange   string                  // Exchange binding
    routingKey string                  // Routing key
    msgType    string                  // Message type
    reflect    *reflect.Value          // Reflection data
    handler    ConsumerHandler         // Handler function
}
```

Reference: `@options.go`

```go
// ✅ CORRECT: Clear struct with tags
type DeliveryMetadata struct {
    MessageID      string                 `json:"message_id"`
    XCount         int64                  `json:"x_count"`
    Type           string                 `json:"type"`
    OriginExchange string                 `json:"origin_exchange"`
    RoutingKey     string                 `json:"routing_key"`
    Headers        map[string]interface{} `json:"headers"`
}
```

---

## Rule: Variable Declarations

**Description**: Variables MUST be declared with appropriate scope and initialization.

**When it applies**: When declaring variables.

**Copilot MUST**:
- Use short variable declarations (`:=`) when possible
- Declare variables close to their use
- Initialize variables when declaring
- Use meaningful variable names

**Copilot MUST NOT**:
- Declare variables far from their use
- Use `var` when short declaration works
- Create variables with unclear purpose

**Example - Correct Variable Declarations**:

Reference: `@tracing.go`

```go
// ✅ CORRECT: Short declaration close to use
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

## Coding Conventions Checklist

When generating code, ensure:

- ✅ PascalCase for exported, camelCase for unexported
- ✅ Descriptive, intention-revealing names
- ✅ Functions under 30 lines
- ✅ Proper import organization (stdlib, then third-party)
- ✅ All exported symbols documented
- ✅ No magic numbers (use named constants)
- ✅ Descriptive error messages (lowercase, no punctuation)
- ✅ Code formatted with gofmt/goimports
- ✅ Early returns to reduce nesting
- ✅ Single responsibility per function
- ✅ Struct fields logically organized
