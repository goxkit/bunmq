---
applyTo: "**/*"
---

# Security Instructions

> **Scope**: Applies to all Go files (`**/*.go`), dependency files (`go.mod`, `go.sum`), and library configuration.

## Rule: Dependency Security

**Description**: All dependencies MUST pass `gosec` security checks and use specific versions.

**When it applies**: When adding dependencies or generating code that uses external packages.

**Copilot MUST**:
- Use specific version constraints in `go.mod` (avoid `*` wildcards)
- Prefer well-maintained packages with active security practices
- Check for known CVEs before adding dependencies
- Run `make sec` for security scanning

**Copilot MUST NOT**:
- Add dependencies with known CVEs
- Use `*` version constraints
- Ignore security warnings
- Use abandoned or unmaintained packages

**Example - Correct Dependency Declaration**:

Reference: `@go.mod`

```go
require (
    github.com/google/uuid v1.6.0            // ✅ Specific version
    github.com/rabbitmq/amqp091-go v1.10.0   // ✅ Specific version, well-maintained
    github.com/sirupsen/logrus v1.9.3        // ✅ Specific version
    go.opentelemetry.io/otel v1.39.0         // ✅ Specific version
    go.opentelemetry.io/otel/trace v1.39.0   // ✅ Specific version
)
```

**Example - Incorrect Dependency Declaration**:

```go
require (
    github.com/some/package v0.0.0-*           // ❌ Wildcard, may include vulnerable versions
    github.com/abandoned/package v1.0.0        // ❌ Abandoned package
    github.com/streadway/amqp v1.0.0          // ❌ Deprecated, use amqp091-go instead
)
```

---

## Rule: Gosec Configuration

**Description**: Configure gosec for appropriate security scanning.

**When it applies**: When setting up or modifying security scanning.

**Copilot MUST**:
- Use the gosec configuration in `.gosec.json`
- Exclude vendor, mock, and test directories from security scans
- Run `make sec` before committing code
- Address all medium and high severity findings

**Copilot MUST NOT**:
- Ignore gosec warnings without justification
- Commit code with security vulnerabilities
- Disable security checks entirely

**Example - Correct Gosec Configuration**:

Reference: `@.gosec.json`

```json
{
  "severity": "medium",
  "confidence": "medium",
  "exclude-dirs": [
    "**/vendor/**",
    "**/mock/**",
    "**/mocks/**",
    "**/testdata/**"
  ],
  "exclude": [
  ]
}
```

---

## Rule: Never Hardcode Connection Strings or Secrets

**Description**: Never hardcode RabbitMQ connection strings, passwords, tokens, or any sensitive data.

**When it applies**: When generating any code that uses configuration or credentials.

**Copilot MUST**:
- Accept connection strings as constructor parameters
- Document required configuration
- Use environment variables for sensitive data in examples
- Never log connection strings or credentials

**Copilot MUST NOT**:
- Commit secrets to version control
- Hardcode passwords, tokens, or API keys
- Log secret values including connection strings
- Store secrets in code comments

**Example - Correct Connection String Handling**:

Reference: `@connection_manager.go`

```go
// ✅ Connection string passed as parameter, never hardcoded
func NewConnectionManager(
    ctx context.Context,
    connectionString string,  // ✅ Passed in, not hardcoded
    appName string,
    options ...ConnectionManagerOption,
) (ConnectionManager, error) {
    // ✅ Store securely, don't log
    cm := &connectionManager{
        connectionString: connectionString,
        appName:          appName,
    }
    return cm, nil
}
```

Reference: `@examples/main.go`

```go
// ✅ Load from environment variable
connectionString := os.Getenv("RABBITMQ_URL")
if connectionString == "" {
    log.Fatal("RABBITMQ_URL environment variable is required")
}
```

**Example - Incorrect Secret Management**:

```go
// ❌ WRONG: Hardcoded connection string
const connectionString = "amqp://admin:password123@localhost:5672/"

// ❌ WRONG: Logging connection string
logrus.WithField("url", connectionString).Info("connecting")

// ❌ WRONG: Secrets in code comments
// Connection: amqp://user:secret@prod.rabbitmq.com:5672/
```

---

## Rule: Input Validation for Messages

**Description**: Always validate all external input before processing.

**When it applies**: When generating code that receives AMQP messages.

**Copilot MUST**:
- Validate message body is valid JSON
- Validate message structure before processing
- Reject invalid input early (fail fast)
- Use appropriate error types for validation failures

**Copilot MUST NOT**:
- Trust external input without validation
- Process messages without unmarshaling validation
- Skip validation for "performance"

**Example - Correct Input Validation**:

Reference: `@dispatcher.go`

```go
// ✅ Validate unmarshaling with proper error handling
func (d *dispatcher) unmarshalDelivery(ctx context.Context, delivery *amqp.Delivery, msgType reflect.Type) (any, error) {
    msg := reflect.New(msgType).Interface()

    // ✅ Validate JSON unmarshaling
    if err := json.Unmarshal(delivery.Body, msg); err != nil {
        logrus.WithContext(ctx).
            WithError(err).
            WithField("body", string(delivery.Body)).
            Error("failed to unmarshal message")
        return nil, fmt.Errorf("failed to unmarshal message: %w", err)
    }

    return msg, nil
}
```

**Example - Validation in Handler**:

```go
// ✅ Handler with validation
func MyHandler(ctx context.Context, msg any, metadata *bunmq.DeliveryMetadata) error {
    orderMsg, ok := msg.(*OrderMessage)
    if !ok {
        return fmt.Errorf("unexpected message type: %T", msg)
    }

    // ✅ Validate required fields
    if orderMsg.OrderID == "" {
        logrus.WithContext(ctx).Error("order_id is required")
        return errors.New("validation failed: order_id is required")
    }

    return nil
}
```

---

## Rule: Thread-Safe Resource Access

**Description**: Ensure concurrent-safe access to shared resources.

**When it applies**: When generating code that manages shared state (connections, channels).

**Copilot MUST**:
- Use sync.Mutex or sync.RWMutex for shared state
- Use RWMutex for read-heavy workloads
- Lock only the minimum necessary scope
- Always unlock in defer statements

**Copilot MUST NOT**:
- Access shared state without synchronization
- Hold locks during I/O operations
- Create potential deadlocks
- Ignore race condition warnings

**Example - Correct Thread-Safe Access**:

Reference: `@connection_manager.go`

```go
type connectionManager struct {
    conn   RMQConnection
    ch     AMQPChannel
    mu     sync.RWMutex  // ✅ Protects conn and ch
    closed bool
}

// ✅ Read lock for checking state
func (cm *connectionManager) IsHealthy() bool {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    if cm.closed {
        return false
    }
    return cm.conn != nil && !cm.conn.IsClosed() && cm.ch != nil
}

// ✅ Write lock for modifying state
func (cm *connectionManager) reconnect() error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    // Safe to modify cm.conn and cm.ch
    // ...
}
```

Reference: `@dispatcher.go`

```go
// ✅ Mutex for consumer map
type dispatcher struct {
    consumers map[string]*consumerDefinition
    mu        sync.RWMutex
}

func (d *dispatcher) RegisterByType(queue string, typE any, handler ConsumerHandler) error {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Safe to modify d.consumers
}
```

---

## Rule: Error Information Leakage Prevention

**Description**: Never expose internal details or sensitive information in errors returned to external systems.

**When it applies**: When generating error handling code.

**Copilot MUST**:
- Log detailed errors internally with full context
- Return domain-specific errors without internal details
- Use predefined error types from errors.go
- Never include connection strings in errors

**Copilot MUST NOT**:
- Expose stack traces externally
- Reveal internal paths or configuration
- Include connection details in error messages

**Example - Correct Error Handling**:

Reference: `@errors.go`

```go
// ✅ Domain-specific error types without internal details
var (
    ErrConnectionClosed     = errors.New("connection is closed")
    ErrChannelClosed        = errors.New("channel is closed")
    ErrFailedToConnect      = errors.New("failed to connect to broker")
    ErrReconnectFailed      = errors.New("reconnection attempts exhausted")
    ErrQueueNotFound        = errors.New("queue definition not found")
    ErrInvalidHandlerType   = errors.New("invalid handler type")
    ErrConsumerAlreadyExists = errors.New("consumer already registered")
)
```

Reference: `@connection_manager.go`

```go
// ✅ Log details internally, return generic error externally
func (cm *connectionManager) reconnect() error {
    // Internal logging with full context
    logrus.WithContext(ctx).
        WithError(err).
        WithField("attempt", attempt).
        Error("connection attempt failed")

    // External return - no connection string exposed
    return ErrReconnectFailed  // ✅ Generic, no internal details
}
```

**Example - Incorrect Error Handling**:

```go
// ❌ WRONG: Exposing connection details
return fmt.Errorf("failed to connect to %s: %v", connectionString, err)

// ❌ WRONG: Internal path exposure
return fmt.Errorf("config error at /etc/rabbitmq/config.json: %v", err)
```

---

## Rule: No Panic in Library Code

**Description**: Library code MUST NOT panic; always return errors.

**When it applies**: When generating any library code.

**Copilot MUST**:
- Return errors instead of panicking
- Validate input and return descriptive errors
- Use recover only in dispatcher for handler panics
- Handle all error cases gracefully

**Copilot MUST NOT**:
- Use panic for error handling
- Allow panics to propagate to library users
- Use panic for validation failures

**Example - Correct Error Handling (No Panic)**:

Reference: `@publisher.go`

```go
// ✅ Return error, never panic
func (p *publisher) Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error {
    if exchange == "" {
        logrus.WithContext(ctx).Error("exchange cannot be empty")
        return fmt.Errorf("exchange cannot be empty")  // ✅ Return error
    }
    return p.publish(ctx, exchange, routingKey, msg, options...)
}
```

Reference: `@dispatcher.go`

```go
// ✅ Recover from handler panics to protect library users
func (d *dispatcher) handleDelivery(ctx context.Context, delivery *amqp.Delivery, def *consumerDefinition) {
    defer func() {
        if r := recover(); r != nil {
            logrus.WithContext(ctx).
                WithField("panic", r).
                Error("handler panicked, message will be nacked")
            _ = delivery.Nack(false, false)
        }
    }()

    // ... handle delivery
}
```

**Example - Incorrect Panic Usage**:

```go
// ❌ WRONG: Panic in library code
func NewPublisher(conn ConnectionManager) Publisher {
    if conn == nil {
        panic("connection manager cannot be nil")  // ❌ Never panic
    }
}

// ✅ CORRECT: Return error instead
func NewPublisher(conn ConnectionManager) (Publisher, error) {
    if conn == nil {
        return nil, fmt.Errorf("connection manager cannot be nil")
    }
    return &publisher{conn: conn}, nil
}
```

---

## Rule: Graceful Shutdown and Resource Cleanup

**Description**: Always implement proper resource cleanup and graceful shutdown.

**When it applies**: When generating code that manages connections, channels, or goroutines.

**Copilot MUST**:
- Use defer for resource cleanup
- Handle shutdown signals (SIGTERM, SIGINT)
- Close channels and connections in correct order
- Cancel contexts to stop goroutines

**Copilot MUST NOT**:
- Leak goroutines
- Leave connections open
- Ignore shutdown signals
- Create orphaned resources

**Example - Correct Graceful Shutdown**:

Reference: `@dispatcher.go`

```go
// ✅ Handle shutdown signals
func (d *dispatcher) Consume(ctx context.Context) error {
    signalCh := make(chan os.Signal, 1)
    signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        select {
        case <-ctx.Done():
            return
        case <-signalCh:
            d.cancel()  // ✅ Cancel context to stop consumers
        }
    }()

    // ...
}
```

Reference: `@connection_manager.go`

```go
// ✅ Proper cleanup order
func (cm *connectionManager) Close() error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    cm.closed = true
    cm.cancel()  // ✅ Cancel context first

    // ✅ Close channel before connection
    if cm.ch != nil {
        _ = cm.ch.Close()
    }

    // ✅ Close connection last
    if cm.conn != nil {
        return cm.conn.Close()
    }

    return nil
}
```

---

## Rule: Secure Channel and Connection Handling

**Description**: Ensure secure handling of AMQP channels and connections.

**When it applies**: When generating code that manages RabbitMQ resources.

**Copilot MUST**:
- Use TLS connections in production (amqps://)
- Validate connection health before operations
- Implement automatic reconnection with backoff
- Close resources in correct order (channel before connection)

**Copilot MUST NOT**:
- Use unencrypted connections in production
- Ignore connection errors
- Leave stale connections
- Share channels across goroutines without synchronization

**Example - Correct Connection Health Check**:

Reference: `@connection_manager.go`

```go
// ✅ Validate connection before returning
func (cm *connectionManager) GetConnection() (RMQConnection, error) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    if cm.closed {
        return nil, ErrConnectionClosed
    }

    if cm.conn == nil || cm.conn.IsClosed() {
        return nil, ErrConnectionClosed
    }

    return cm.conn, nil
}
```

---

## Security Checklist

When generating bunmq code, ensure:

- ✅ No hardcoded connection strings or secrets
- ✅ All message input validated (JSON unmarshaling)
- ✅ Thread-safe access to shared resources (sync.Mutex)
- ✅ Errors sanitized (no connection strings exposed)
- ✅ Dependencies scanned for vulnerabilities (`make sec`)
- ✅ No sensitive data in logs
- ✅ No panics in library code (return errors)
- ✅ Graceful shutdown implemented
- ✅ Resources cleaned up properly (defer)
- ✅ Channel closed before connection

---

## Rule: Gosec Integration

**Description**: Run gosec as part of the development workflow.

**When it applies**: When developing or modifying code.

**Copilot MUST**:
- Run `make sec` before committing
- Fix all security findings
- Use `make sec-report` for detailed reports

**Commands**:

Reference: `@Makefile`

```makefile
# Run security scanner
make sec

# Generate security report
make sec-report
```

---

## Anti-Patterns (FORBIDDEN)

**Copilot MUST NEVER generate**:

```go
// ❌ Hardcoded connection string
conn, _ := amqp.Dial("amqp://user:pass@localhost:5672/")

// ❌ Logging connection string
logrus.Info("connecting to " + connectionString)

// ❌ Panic in library
if ch == nil {
    panic("channel is nil")
}

// ❌ Ignoring errors
ch, _ := conn.Channel()

// ❌ Unprotected shared state
func (d *dispatcher) Register(queue string, h Handler) {
    d.consumers[queue] = h  // ❌ No mutex protection
}

// ❌ Leaked goroutine
go func() {
    for msg := range msgs {
        process(msg)
    }
}()  // ❌ No way to stop this goroutine
```

---

## Positive Security Patterns (REQUIRED)

**Copilot MUST generate**:

```go
// ✅ Connection string from parameter
func NewConnectionManager(ctx context.Context, connStr string, appName string) (ConnectionManager, error)

// ✅ Protected shared state
cm.mu.Lock()
defer cm.mu.Unlock()
cm.consumers[queue] = handler

// ✅ Error handling
ch, err := conn.Channel()
if err != nil {
    return nil, fmt.Errorf("failed to open channel: %w", err)
}

// ✅ Resource cleanup
defer func() {
    if ch != nil {
        _ = ch.Close()
    }
}()

// ✅ Cancellable goroutine
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-msgs:
            process(msg)
        }
    }
}()
```
