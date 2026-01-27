---
applyTo: "**/*.go"
---

# Go Language Instructions

## Rule: Zero Linter Warnings

**Description**: All code MUST pass `golangci-lint` without warnings.

**When it applies**: When generating any Go code.

**Copilot MUST**:
- Generate code that passes `golangci-lint`
- Ensure code passes `go vet` and `staticcheck`
- Use idiomatic Go patterns
- Follow Go naming conventions

**Copilot MUST NOT**:
- Generate code that produces linter warnings
- Use `//nolint` comments without justification
- Disable linter rules at package level
- Ignore linter suggestions

**Tool Reference**: See `@Makefile` for lint configuration.

```bash
# Run linter
make lint
```

---

## Rule: Format Code with gofmt

**Description**: All code MUST be formatted according to `gofmt` standards.

**When it applies**: When generating any Go code.

**Copilot MUST**:
- Format code according to `gofmt` standards
- Organize imports with `goimports`
- Use tabs for indentation
- Keep lines under 120 characters when possible

**Copilot MUST NOT**:
- Use non-standard formatting
- Mix tabs and spaces
- Generate lines exceeding 120 characters unnecessarily

---

## Rule: Enforce gosec for Security

**Description**: All code MUST pass `gosec` security checks.

**When it applies**: When generating code or adding dependencies.

**Copilot MUST**:
- Use specific version constraints in `go.mod`
- Prefer well-maintained packages
- Check for known CVEs before adding dependencies
- Avoid security anti-patterns

**Copilot MUST NOT**:
- Add dependencies with known CVEs
- Use `*` version constraints
- Ignore security warnings
- Use unsafe patterns (SQL injection, path traversal, etc.)

**Example - Correct Dependency Declaration**:

Reference: `@go.mod`

```go
require (
    github.com/google/uuid v1.6.0               // ✅ Specific version
    github.com/rabbitmq/amqp091-go v1.10.0      // ✅ Specific version
    github.com/sirupsen/logrus v1.9.3           // ✅ Specific version
    go.opentelemetry.io/otel v1.39.0            // ✅ Specific version
)
```

**Tool Reference**: See `@Makefile` for security scanning.

```bash
# Run security scanner
make sec

# Generate security reports
make sec-report
```

---

## Rule: Error Handling

**Description**: Always handle errors explicitly, never ignore them.

**When it applies**: When handling any error condition.

**Copilot MUST**:
- Always check error returns
- Use domain-specific error types when appropriate
- Wrap errors with context: `fmt.Errorf("action: %w", err)`
- Use `errors.Is()` and `errors.As()` for error checking
- Log errors before returning them

**Copilot MUST NOT**:
- Use `panic()` for recoverable errors (library code should NEVER panic)
- Return raw string errors
- Ignore errors with `_`
- Lose error context

**Example - Correct Error Handling**:

Reference: `@publisher.go`

```go
// ✅ CORRECT: Validate input and return clear error
func (p *publisher) Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error {
    if exchange == "" {
        logrus.WithContext(ctx).Error("exchange cannot be empty")
        return fmt.Errorf("exchange cannot be empty")
    }
    return p.publish(ctx, exchange, routingKey, msg, options...)
}
```

Reference: `@errors.go`

```go
// ✅ CORRECT: Domain-specific error variables
var (
    ErrQueueNotDefined       = errors.New("queue not defined in topology")
    ErrHandlerAlreadyExists  = errors.New("handler already exists for this queue")
    ErrConnectionClosed      = errors.New("connection is closed")
)
```

**Example - Incorrect Error Handling**:

```go
// ❌ WRONG: Ignoring error
conn, _ := amqp.Dial(url)

// ❌ WRONG: Panic in library code
if err != nil {
    panic(err)
}

// ❌ WRONG: Generic error without context
return errors.New("failed")
```

---

## Rule: Context Management

**Description**: Always accept `context.Context` as the first parameter for I/O operations.

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

**Example - Correct Context Usage**:

Reference: `@publisher.go`

```go
// ✅ CORRECT: Context as first parameter
func (p *publisher) Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error {
    // Use context for logging
    logrus.WithContext(ctx).Debug("publishing message")

    // Propagate context
    return p.publish(ctx, exchange, routingKey, msg, options...)
}
```

Reference: `@dispatcher.go`

```go
// ✅ CORRECT: Context for tracing
func (d *dispatcher) handleDelivery(ctx context.Context, delivery amqp.Delivery, def *ConsumerDefinition) error {
    ctx, span := d.tracer.Start(ctx, "dispatcher.handleDelivery")
    defer span.End()

    // Context propagated to handler
    return def.handler(ctx, msg, metadata)
}
```

---

## Rule: Interface Design

**Description**: Define interfaces near consumers, keep them small and focused.

**When it applies**: When creating new interfaces or dependencies.

**Copilot MUST**:
- Keep interfaces small and focused (Interface Segregation Principle)
- Define interfaces where they are used (consumer side)
- Use interfaces for dependency injection
- Document interface methods clearly

**Copilot MUST NOT**:
- Create large interfaces that violate ISP
- Put interface definitions far from consumers
- Create interfaces with single implementations (unless for testing/mocking)

**Example - Correct Interface Design**:

Reference: `@connection.go`

```go
// ✅ CORRECT: Small, focused interface
type RMQConnection interface {
    Channel() (*amqp.Channel, error)
    ConnectionState() tls.ConnectionState
    IsClosed() bool
    Close() error
    NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
}
```

Reference: `@publisher.go`

```go
// ✅ CORRECT: Clear, focused interface
type Publisher interface {
    Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error
    PublishDeadline(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error
    PublishQueue(ctx context.Context, queue string, msg any, options ...*Option) error
    PublishQueueDeadline(ctx context.Context, queue string, msg any, options ...*Option) error
}
```

---

## Rule: Struct Organization

**Description**: Organize struct fields logically and provide constructors.

**When it applies**: When creating new structs.

**Copilot MUST**:
- Organize fields: exported first, then unexported
- Group related fields together
- Provide constructors for structs that need initialization
- Use field tags for JSON, validation when needed

**Example - Correct Struct Organization**:

Reference: `@dispatcher.go`

```go
// ✅ CORRECT: Logical field organization
type dispatcher struct {
    manager             ConnectionManager                    // Core dependency
    queueDefinitions    map[string]*QueueDefinition          // Configuration
    consumersDefinition map[string][]*ConsumerDefinition     // State
    tracer              trace.Tracer                         // Observability
    signalCh            chan os.Signal                       // Signal handling
    mu                  sync.RWMutex                         // Synchronization
}

// ✅ CORRECT: Constructor with proper initialization
func NewDispatcher(manager ConnectionManager, definitions []*QueueDefinition) Dispatcher {
    d := &dispatcher{
        manager:             manager,
        queueDefinitions:    make(map[string]*QueueDefinition),
        consumersDefinition: make(map[string][]*ConsumerDefinition),
        tracer:              otel.Tracer("bunmq-dispatcher"),
        signalCh:            make(chan os.Signal, 1),
    }
    for _, def := range definitions {
        d.queueDefinitions[def.name] = def
    }
    return d
}
```

---

## Rule: Import Organization

**Description**: Organize imports in groups: standard library, third-party, internal.

**When it applies**: When generating import statements.

**Copilot MUST**:
- Group imports: standard library, third-party
- Separate groups with blank lines
- Order imports alphabetically within groups
- Use aliased imports when needed (e.g., `amqp "github.com/rabbitmq/amqp091-go"`)

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

## Rule: Function Size and Complexity

**Description**: Keep functions under 30 lines (excluding comments and blank lines).

**When it applies**: When generating any function.

**Copilot MUST**:
- Keep functions under 30 lines
- Extract complex logic to helper functions
- Use early returns (guard clauses) to reduce nesting
- Ensure single responsibility per function

**Copilot MUST NOT**:
- Create functions over 50 lines
- Allow deep nesting (> 3 levels)
- Mix multiple concerns in one function

---

## Rule: Defer for Resource Cleanup

**Description**: Use `defer` for resource cleanup to ensure proper release.

**When it applies**: When managing resources that need cleanup (connections, channels, files, locks).

**Copilot MUST**:
- Use `defer` for cleanup immediately after resource acquisition
- Use `defer` for mutex unlocking
- Handle errors from deferred close operations when critical

**Copilot MUST NOT**:
- Skip cleanup for resources
- Use defer in performance-critical tight loops (evaluate case by case)

**Example - Correct Defer Usage**:

Reference: `@dispatcher.go`

```go
// ✅ CORRECT: Defer for span closure
func (d *dispatcher) handleDelivery(ctx context.Context, delivery amqp.Delivery, def *ConsumerDefinition) error {
    ctx, span := d.tracer.Start(ctx, "dispatcher.handleDelivery")
    defer span.End()

    // ...
}

// ✅ CORRECT: Defer for mutex unlock
func (d *dispatcher) RegisterByType(queue string, typE any, handler ConsumerHandler) error {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Safe to modify shared state
    d.consumersDefinition[queue] = append(d.consumersDefinition[queue], def)
    return nil
}
```

---

## Quality Gates Summary

All Go code MUST pass these quality gates (enforced in CI/CD):

| Tool | Command | Description |
|------|---------|-------------|
| golangci-lint | `make lint` | Zero warnings |
| go vet | automatic | No issues |
| gosec | `make sec` | No security vulnerabilities |
| go test | `make test` | All tests passing |
| coverage | `make test-coverage-threshold` | 90%+ coverage |

---

## Go Idioms Checklist

When generating code, ensure:

- ✅ Zero linter warnings
- ✅ Properly formatted with gofmt
- ✅ No security vulnerabilities (gosec)
- ✅ All errors handled explicitly
- ✅ Context as first parameter for I/O functions
- ✅ Interfaces small and focused
- ✅ Imports properly organized
- ✅ Functions under 30 lines
- ✅ Defer used for resource cleanup
- ✅ No panic in library code
