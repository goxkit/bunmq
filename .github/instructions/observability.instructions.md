---
applyTo: "**/*.go"
---

# Observability Instructions

## Rule: Use Structured Logging with Logrus

**Description**: All logging MUST use Logrus with structured fields, never string concatenation.

**When it applies**: When generating any logging code.

**Copilot MUST**:
- Use Logrus for all logging operations
- Use structured logging with fields: `logrus.WithFields()`
- Include context using `logrus.WithContext(ctx)`
- Use appropriate log levels: Debug, Info, Warn, Error

**Copilot MUST NOT**:
- Use `fmt.Printf`, `log.Print`, or `log.Println`
- Concatenate strings in log messages
- Log without context
- Use `Fatal` or `Panic` except in main/init

**Example - Correct Structured Logging**:

Reference: `@internal/services/authorized/service.go`

```go
func (as *authorizedService) Process(
    ctx context.Context,
    customer *configuration.CustomerConfig,
    msg *authorizer.AuthorizedMessage,
) error {
    // ✅ Structured logging with context
    if msg == nil {
        logrus.WithContext(ctx).Error("authorizedMessage cannot be nil")
        return errors.ErrUnformattedMessage
    }
    
    wasProcessed, err := as.eventsRepo.WasProcessed(ctx, customer, ...)
    if err != nil {
        // ✅ Structured logging with error and fields
        logrus.WithContext(ctx).
            WithError(err).
            WithFields(msg.LogFields()).
            Error("failed to check existing was processed in database")
        return err
    }
    
    if wasProcessed {
        // ✅ Info level for expected behavior
        logrus.WithContext(ctx).
            WithFields(msg.LogFields()).
            Info("was processed already exists for the provided NSU, MerchantID, and AuthCode. Skipping insertion.")
        return nil
    }
}
```

**Example - Incorrect Logging**:

```go
// ❌ WRONG: String concatenation
fmt.Printf("Processing message: %+v\n", msg)

// ❌ WRONG: No context
logrus.Error("error occurred")

// ❌ WRONG: Generic message
log.Println("failed")
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
- Use `Fatal` or `Panic` in library code
- Log at `Debug` level in production hot paths

**Example - Correct Log Levels**:

```go
// ✅ Info: Normal operation
logrus.WithContext(ctx).
    WithFields(msg.LogFields()).
    Info("authorized transaction processed successfully")

// ✅ Info: Expected business event (duplicate)
logrus.WithContext(ctx).
    WithFields(msg.LogFields()).
    Info("was processed already exists for the provided NSU, MerchantID, and AuthCode. Skipping insertion.")

// ✅ Warn: Recoverable issue
logrus.WithContext(ctx).
    WithError(err).
    WithField("retry_count", attempt).
    Warn("failed to publish message, retrying")

// ✅ Error: Failure that needs attention
logrus.WithContext(ctx).
    WithError(err).
    WithFields(msg.LogFields()).
    Error("failed to check existing was processed in database")
```

---

## Rule: Include Structured Fields

**Description**: Always include relevant structured fields in log entries.

**When it applies**: When generating logging statements.

**Copilot MUST**:
- Use `WithFields()` or `WithField()` for structured data
- Include identifiers (IDs, keys, transaction numbers) in log fields
- Use message's `LogFields()` method when available
- Include error context with `WithError(err)`

**Copilot MUST NOT**:
- Embed data in log message strings
- Log without relevant context fields
- Log sensitive data (see security instructions)

**Example - Correct Structured Fields**:

Reference: `@internal/services/authorized/service.go`

```go
// ✅ Using message's LogFields() method
logrus.WithContext(ctx).
    WithError(err).
    WithFields(msg.LogFields()). // ✅ Structured fields from message
    Error("failed to check existing was processed in database")

// ✅ Manual fields when needed
logrus.WithContext(ctx).
    WithFields(logrus.Fields{
        "nsu": msg.SystemTraceAuditNumber,
        "merchant_id": msg.MerchantID,
        "transaction_type": msg.TransactionType,
        "product": msg.Product,
    }).
    Info("processing authorized transaction")
```

---

## Rule: OpenTelemetry Tracing

**Description**: Create spans for significant operations and propagate context.

**When it applies**: When generating code that performs I/O or business logic operations.

**Copilot MUST**:
- Create spans for database queries, external API calls, message publishing/consuming
- Propagate context through all layers
- Record errors in spans using `span.RecordError(err)`
- Set span status: `span.SetStatus(codes.Ok, "")` or `span.SetStatus(codes.Error, err.Error())`

**Copilot MUST NOT**:
- Create spans without proper context propagation
- Ignore errors in spans
- Create spans for trivial operations

**Example - Correct Tracing**:

Reference: `@internal/repositories/events_repository.go`

```go
func (pr *clearingEventsRepository) WasProcessed(
    ctx context.Context,
    customer *configuration.CustomerConfig,
    nsu, mid, authTransactionDate, authCode string,
) (bool, error) {
    // ✅ Create span for database operation
    ctx, span := pr.tracer.Start(ctx, "clearingEventsRepository.WasProcessed")
    defer span.End()
    
    db, err := pr.multiDBManager.GetRODB(customer)
    if err != nil {
        logrus.WithContext(ctx).WithError(err).Error("failed to get read-only db connection")
        // ✅ Record error in span
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return false, errors.ErrDatabaseConnection
    }
    
    // ... query logic ...
    
    // ✅ Set success status
    span.SetStatus(codes.Ok, "")
    return false, nil
}
```

---

## Rule: Context Propagation

**Description**: Always propagate `context.Context` through all function calls.

**When it applies**: When generating any function that makes external calls.

**Copilot MUST**:
- Accept `context.Context` as the first parameter
- Propagate context to all called functions
- Use context for tracing, logging, and cancellation
- Create spans using context

**Copilot MUST NOT**:
- Create new contexts without reason
- Store context in structs
- Ignore context cancellation

**Example - Correct Context Propagation**:

Reference: `@internal/transport/rmq/consumers/authorized.go`

```go
func (c *authorizedConsumer) Handle(ctx context.Context, delivery *amqp091.Delivery) error {
    // ✅ Context propagated from consumer
    logrus.WithContext(ctx).Debug("received message in AuthorizedConsumer")
    
    // ✅ Context propagated to service
    return c.authorizerService.Process(ctx, customer, &authorizedMessage)
}
```

---

## Rule: Never Log Sensitive Data

**Description**: Never log passwords, tokens, API keys, credit card numbers, or PII.

**When it applies**: When generating any logging code.

**Copilot MUST**:
- Redact or hash sensitive fields before logging
- Use structured fields to exclude sensitive data
- Log identifiers only, not full payloads
- Validate that no secrets are in log messages

**Copilot MUST NOT**:
- Log full message payloads without filtering
- Log credentials or authentication tokens
- Log customer PII without consent
- Embed sensitive data in log strings

**Example - Correct Logging (No Sensitive Data)**:

```go
// ✅ Log identifiers, not sensitive data
logrus.WithContext(ctx).
    WithFields(logrus.Fields{
        "nsu": msg.SystemTraceAuditNumber, // ✅ Safe identifier
        "merchant_id": msg.MerchantID,      // ✅ Safe identifier
        "transaction_type": msg.TransactionType, // ✅ Safe
    }).
    Info("processing transaction")

// ✅ Log error without exposing sensitive details
logrus.WithContext(ctx).
    WithError(err).
    WithField("operation", "process_transaction").
    Error("failed to process transaction")
```

**Example - Incorrect Logging (Sensitive Data)**:

```go
// ❌ WRONG: Logging full message (may contain sensitive data)
logrus.WithContext(ctx).WithField("message", msg).Debug("received message")

// ❌ WRONG: Logging credentials
logrus.WithContext(ctx).WithField("password", password).Info("authenticating")

// ❌ WRONG: Logging tokens
logrus.WithContext(ctx).WithField("token", authToken).Debug("using token")
```

---

## Rule: Error Logging Best Practices

**Description**: Always log errors with full context before returning them.

**When it applies**: When handling any error condition.

**Copilot MUST**:
- Log errors before returning them
- Include full context: error message, identifiers, operation context
- Use `WithError(err)` to include error in structured logs
- Use appropriate log level (Error for failures)

**Copilot MUST NOT**:
- Log errors without context
- Swallow errors silently
- Log the same error multiple times

**Example - Correct Error Logging**:

Reference: `@internal/services/authorized/service.go`

```go
wasProcessed, err := as.eventsRepo.WasProcessed(ctx, customer, msg.SystemTraceAuditNumber, msg.MerchantID, msg.TransactionDate, msg.AuthorizationCode)
if err != nil {
    // ✅ Log error with full context before returning
    logrus.WithContext(ctx).
        WithError(err).                    // ✅ Include error
        WithFields(msg.LogFields()).        // ✅ Include message fields
        Error("failed to check existing was processed in database")
    return err
}
```

---

## Observability Checklist

When generating code, ensure:

- ✅ All errors are logged with context
- ✅ Structured logging used (no string concatenation)
- ✅ Context propagated through all layers
- ✅ Spans created for significant operations
- ✅ No sensitive data in logs
- ✅ Appropriate log levels used
- ✅ Correlation IDs included in logs
