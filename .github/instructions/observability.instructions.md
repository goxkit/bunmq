---
applyTo: "**/*.go"
---

# Observability Instructions

## Rule: Use Structured Logging with Logrus

**Description**: All logging MUST use Logrus with structured fields, never string concatenation.

**When it applies**: When generating any logging code.

**Copilot MUST**:
- Use Logrus for all logging operations
- Use structured logging with fields: `logrus.WithFields()` or `logrus.WithField()`
- Include context using `logrus.WithContext(ctx)`
- Use appropriate log levels: Debug, Info, Warn, Error

**Copilot MUST NOT**:
- Use `fmt.Printf`, `log.Print`, or `log.Println`
- Concatenate strings in log messages
- Log without context
- Use `Fatal` or `Panic` in library code

**Example - Correct Structured Logging**:

Reference: `@publisher.go`

```go
// ✅ Structured logging with context
func (p *publisher) Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error {
    if exchange == "" {
        logrus.WithContext(ctx).Error("exchange cannot be empty")
        return fmt.Errorf("exchange cannot be empty")
    }
    return p.publish(ctx, exchange, routingKey, msg, options...)
}
```

Reference: `@dispatcher.go`

```go
// ✅ Structured logging with fields
logrus.WithContext(ctx).
    WithField("queue", queue).
    WithField("consumer_tag", consumerTag).
    Info("started consuming from queue")
```

**Example - Incorrect Logging**:

```go
// ❌ WRONG: String concatenation
fmt.Printf("Publishing to exchange: %s\n", exchange)

// ❌ WRONG: No context
logrus.Error("error occurred")

// ❌ WRONG: Generic message
log.Println("failed")

// ❌ WRONG: Printf-style logging
logrus.Infof("processing message %v", msg)  // Use WithField instead
```

---

## Rule: Log Levels

**Description**: Use log levels appropriately based on severity and context.

**When it applies**: When generating logging statements.

**Copilot MUST**:
- Use **Debug** for detailed diagnostic information (development only)
- Use **Info** for general informational messages about normal operation
- Use **Warn** for potentially harmful situations that are recoverable
- Use **Error** for error events that need attention

**Copilot MUST NOT**:
- Use `Error` for expected conditions (use `Info` or `Warn`)
- Use `Fatal` or `Panic` in library code (never in bunmq)
- Log at `Debug` level in production hot paths

**Example - Correct Log Levels**:

Reference: `@dispatcher.go`

```go
// ✅ Debug: Detailed diagnostic information
logrus.WithContext(ctx).
    WithField("delivery_tag", delivery.DeliveryTag).
    Debug("received delivery")

// ✅ Info: Normal operation events
logrus.WithContext(ctx).
    WithField("queue", queue).
    Info("consumer registered successfully")

// ✅ Warn: Recoverable issues
logrus.WithContext(ctx).
    WithError(err).
    WithField("retry_count", attempt).
    Warn("failed to publish message, retrying")

// ✅ Error: Failures that need attention
logrus.WithContext(ctx).
    WithError(err).
    Error("failed to connect to RabbitMQ")
```

---

## Rule: Include Structured Fields

**Description**: Always include relevant structured fields in log entries.

**When it applies**: When generating logging statements.

**Copilot MUST**:
- Use `WithFields()` or `WithField()` for structured data
- Include relevant identifiers (queue names, exchange names, routing keys)
- Include error context with `WithError(err)`
- Include timing information when relevant

**Copilot MUST NOT**:
- Embed data in log message strings
- Log without relevant context fields
- Log sensitive data (credentials, tokens)

**Example - Correct Structured Fields**:

Reference: `@dispatcher.go`

```go
// ✅ Multiple relevant fields
logrus.WithContext(ctx).
    WithFields(logrus.Fields{
        "queue":       queueName,
        "exchange":    exchange,
        "routing_key": routingKey,
        "consumer_tag": consumerTag,
    }).
    Info("binding consumer to queue")

// ✅ Error with context
logrus.WithContext(ctx).
    WithError(err).
    WithField("queue", queue).
    Error("failed to declare queue")
```

---

## Rule: OpenTelemetry Tracing

**Description**: Create spans for significant operations and propagate trace context.

**When it applies**: When generating code that performs I/O or significant operations.

**Copilot MUST**:
- Create spans for message publishing and consuming
- Propagate trace context through AMQP headers
- Record errors in spans using `span.RecordError(err)`
- Set span status: `span.SetStatus(codes.Ok, "")` or `span.SetStatus(codes.Error, err.Error())`
- Use `defer span.End()` to ensure span closure

**Copilot MUST NOT**:
- Create spans without proper context propagation
- Ignore errors in spans
- Create spans for trivial operations
- Forget to end spans

**Example - Correct Tracing**:

Reference: `@tracing.go`

```go
// ✅ AMQPHeader implements TextMapCarrier for trace propagation
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

Reference: `@dispatcher.go`

```go
// ✅ Create span for message handling
func (d *dispatcher) handleDelivery(ctx context.Context, delivery amqp.Delivery, def *ConsumerDefinition) error {
    ctx, span := d.tracer.Start(ctx, "dispatcher.handleDelivery")
    defer span.End()

    // ... processing logic ...

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.SetStatus(codes.Ok, "")
    return nil
}
```

---

## Rule: Trace Context Propagation in AMQP

**Description**: Propagate OpenTelemetry trace context through AMQP message headers.

**When it applies**: When publishing or consuming AMQP messages.

**Copilot MUST**:
- Inject trace context into AMQP headers when publishing
- Extract trace context from AMQP headers when consuming
- Use the `AMQPPropagator` composite propagator
- Support both TraceContext and Baggage propagation

**Copilot MUST NOT**:
- Skip trace propagation in messages
- Use custom header names (use standard W3C trace headers)
- Lose trace context between services

**Example - Correct Trace Propagation**:

Reference: `@tracing.go`

```go
// ✅ Composite propagator for AMQP
var AMQPPropagator = propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{},
    propagation.Baggage{},
)

// ✅ Inject trace context when publishing
func InjectTraceContext(ctx context.Context, headers amqp.Table) {
    carrier := AMQPHeader(headers)
    AMQPPropagator.Inject(ctx, carrier)
}

// ✅ Extract trace context when consuming
func ExtractTraceContext(ctx context.Context, headers amqp.Table) context.Context {
    carrier := AMQPHeader(headers)
    return AMQPPropagator.Extract(ctx, carrier)
}
```

---

## Rule: Context Propagation

**Description**: Always propagate `context.Context` through all function calls.

**When it applies**: When generating any function that performs operations.

**Copilot MUST**:
- Accept `context.Context` as the first parameter
- Propagate context to all called functions
- Use context for tracing, logging, and cancellation
- Extract/inject trace context from/to AMQP headers

**Copilot MUST NOT**:
- Create new contexts without reason
- Store context in structs
- Ignore context cancellation
- Pass `nil` context

**Example - Correct Context Propagation**:

Reference: `@publisher.go`

```go
// ✅ Context as first parameter, propagated throughout
func (p *publisher) Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error {
    // Context used for logging
    logrus.WithContext(ctx).Debug("publishing message")

    // Context propagated to internal method
    return p.publish(ctx, exchange, routingKey, msg, options...)
}
```

---

## Rule: Never Log Sensitive Data

**Description**: Never log passwords, tokens, API keys, or sensitive message content.

**When it applies**: When generating any logging code.

**Copilot MUST**:
- Log identifiers only, not full message payloads
- Redact or exclude sensitive fields
- Use structured fields to control what is logged
- Review log statements for sensitive data

**Copilot MUST NOT**:
- Log full message bodies (may contain sensitive data)
- Log credentials or authentication tokens
- Log connection strings with passwords
- Use `%+v` or `%#v` on message objects

**Example - Correct Logging (No Sensitive Data)**:

```go
// ✅ Log identifiers, not content
logrus.WithContext(ctx).
    WithFields(logrus.Fields{
        "exchange":    exchange,
        "routing_key": routingKey,
        "message_id":  messageID,
    }).
    Debug("publishing message")

// ❌ WRONG: Logging full message
logrus.WithContext(ctx).WithField("message", msg).Debug("received message")

// ❌ WRONG: Logging connection URL (contains password)
logrus.WithField("url", amqpURL).Info("connecting")
```

---

## Rule: Error Logging Best Practices

**Description**: Always log errors with full context before returning them.

**When it applies**: When handling any error condition.

**Copilot MUST**:
- Log errors before returning them
- Include full context: operation, identifiers, error
- Use `WithError(err)` to include error in structured logs
- Use Error level for failures

**Copilot MUST NOT**:
- Log errors without context
- Swallow errors silently
- Log the same error multiple times (log once, at the source)

**Example - Correct Error Logging**:

Reference: `@publisher.go`

```go
// ✅ Log error with context at point of detection
if exchange == "" {
    logrus.WithContext(ctx).Error("exchange cannot be empty")
    return fmt.Errorf("exchange cannot be empty")
}

// ✅ Log error with operation context
if err != nil {
    logrus.WithContext(ctx).
        WithError(err).
        WithField("exchange", exchange).
        Error("failed to publish message")
    return err
}
```

---

## Observability Checklist

When generating code, ensure:

- ✅ Structured logging used (no string concatenation)
- ✅ Context propagated through all layers
- ✅ All errors logged with context before returning
- ✅ Spans created for I/O operations (publish, consume)
- ✅ Trace context propagated through AMQP headers
- ✅ No sensitive data in logs
- ✅ Appropriate log levels used
- ✅ `WithContext(ctx)` always used with Logrus
- ✅ Spans properly ended with `defer span.End()`
