---
applyTo: "**/*.go"
---

# Architecture Instructions

## Rule: Enforce Layered Architecture

**Description**: Code MUST be organized into distinct layers with clear boundaries.

**When it applies**: When generating or modifying code in any layer.

**Copilot MUST**:
- Separate transport, service, and repository concerns
- Ensure dependencies flow inward: Transport → Service → Repository
- Keep business logic in service layer only
- Keep data access in repository layer only
- Keep orchestration in transport layer only

**Copilot MUST NOT**:
- Put business logic in message handlers/consumers
- Allow data access code in service layer
- Create direct dependencies between non-adjacent layers
- Mix concerns within a single layer

**Example - Correct Pattern**:

Reference: `@internal/transport/rmq/consumers/authorized.go`

```go
// Transport layer - orchestration only
func (c *authorizedConsumer) Handle(ctx context.Context, delivery *amqp091.Delivery) error {
    // 1. Unmarshal message
    var authorizedMessage authorizer.AuthorizedMessage
    if err := json.Unmarshal(delivery.Body, &authorizedMessage); err != nil {
        return errors.ErrFailedToUnmarshalJSON
    }
    
    // 2. Validate
    if err := authorizedMessage.Validate(); err != nil {
        return errors.ErrUnformattedMessage
    }
    
    // 3. Get customer config
    customer, err := configs.GetCustomerFromConfigTool(ctx, c.configTool, authorizedMessage.ClientKey)
    
    // 4. Delegate to service (business logic)
    return c.authorizerService.Process(ctx, customer, &authorizedMessage)
}
```

Reference: `@internal/services/authorized/service.go`

```go
// Service layer - business logic
func (as *authorizedService) Process(ctx context.Context, customer *configuration.CustomerConfig, msg *authorizer.AuthorizedMessage) error {
    // Business logic: check idempotency
    wasProcessed, err := as.eventsRepo.WasProcessed(ctx, customer, msg.SystemTraceAuditNumber, msg.MerchantID, msg.TransactionDate, msg.AuthorizationCode)
    if wasProcessed {
        return nil // Skip duplicate
    }
    
    // Business logic: select operation
    op, err := as.operationFactory.GetOperation(msg)
    if err != nil {
        return err
    }
    
    // Delegate to operation
    return op.Process(ctx, customer, msg)
}
```

**Example - Incorrect Pattern**:

```go
// ❌ WRONG: Business logic in transport layer
func (c *authorizedConsumer) Handle(ctx context.Context, delivery *amqp091.Delivery) error {
    var msg authorizer.AuthorizedMessage
    json.Unmarshal(delivery.Body, &msg)
    
    // ❌ Business logic directly in consumer
    db, _ := sql.Open(...)
    var exists bool
    db.QueryRow("SELECT EXISTS(...)").Scan(&exists)
    if exists {
        return nil
    }
    
    // ❌ Direct database access in consumer
    db.Exec("INSERT INTO ...")
    
    return nil
}
```

---

## Rule: Dependency Injection via Constructor

**Description**: All dependencies MUST be injected via constructor parameters.

**When it applies**: When creating new services, repositories, or handlers.

**Copilot MUST**:
- Accept dependencies as constructor parameters
- Use interfaces for all dependencies
- Return interfaces, not concrete types
- Wire dependencies in `cmd/container.go`

**Copilot MUST NOT**:
- Create dependencies inside methods
- Use global static instances
- Hardcode concrete implementations
- Hide dependency creation inside constructors

**Example - Correct Pattern**:

Reference: `@internal/services/authorized/service.go`

```go
type authorizedService struct {
    operationFactory interfaces.OperationFactory[authorizer.AuthorizedMessage]
    eventsRepo       interfaces.EventsRepository
}

func NewAuthorizedService(
    operationFactory interfaces.OperationFactory[authorizer.AuthorizedMessage],
    eventsRepo interfaces.EventsRepository,
) interfaces.AuthorizerService {
    return &authorizedService{
        operationFactory: operationFactory,
        eventsRepo:       eventsRepo,
    }
}
```

**Example - Incorrect Pattern**:

```go
// ❌ WRONG: Creating dependencies internally
type badService struct {
    repo *eventsRepository
}

func NewBadService() *badService {
    // ❌ Hidden dependency creation
    db, _ := sql.Open(...)
    return &badService{
        repo: newEventsRepository(db), // ❌ Cannot test, cannot swap implementation
    }
}
```

---

## Rule: Single Responsibility Principle

**Description**: Each struct, function, and module MUST have ONE clear responsibility.

**When it applies**: When generating any new code.

**Copilot MUST**:
- Keep functions under 30 lines (excluding comments)
- Ensure each struct has a single purpose
- Extract complex logic to helper functions
- Use early returns to reduce nesting

**Copilot MUST NOT**:
- Create god objects or god services
- Mix unrelated responsibilities
- Create functions over 50 lines
- Allow deep nesting (> 3 levels)

**Example - Correct Pattern**:

Reference: `@internal/services/authorized/operations/visa/credit.go`

```go
// Single responsibility: Process Visa credit transactions
type visaCredit struct {
    publisher messaging.Producer
}

func (vc *visaCredit) Process(ctx context.Context, customer *configuration.CustomerConfig, msg *authorizer.AuthorizedMessage) error {
    // 1. Convert message (single responsibility)
    enrichMsg, err := converters.AuthorizedToVisaEnrich(msg)
    if err != nil {
        return errors.ErrUnformattedMessage
    }
    
    // 2. Publish (single responsibility)
    return vc.publish(ctx, enrichMsg)
}

// Separate method for publishing (single responsibility)
func (v *visaCredit) publish(ctx context.Context, msg *messages.VisaEnrichMessage) error {
    // Publishing logic
}
```

---

## Rule: Dependency Inversion (Interfaces over Concretions)

**Description**: Code MUST depend on interfaces, not concrete implementations.

**When it applies**: When defining dependencies or creating new components.

**Copilot MUST**:
- Define interfaces in `internal/interfaces/` or near consumers
- Use interfaces for all external dependencies
- Return interfaces from constructors
- Mock interfaces in tests

**Copilot MUST NOT**:
- Depend on concrete types when interfaces exist
- Put interface definitions in implementation files
- Create tight coupling to specific implementations

**Example - Correct Pattern**:

Reference: `@internal/services/interfaces/authorized_service.go`

```go
// Interface defined near consumers
type AuthorizerService interface {
    Process(ctx context.Context, customer *configuration.CustomerConfig, msg *authorizer.AuthorizedMessage) error
}
```

Reference: `@internal/services/authorized/service.go`

```go
// Implementation depends on interfaces
type authorizedService struct {
    operationFactory interfaces.OperationFactory[authorizer.AuthorizedMessage] // Interface
    eventsRepo       interfaces.EventsRepository // Interface
}
```

---

## Rule: Factory Pattern for Operation Selection

**Description**: Use factory pattern to select operations based on message properties.

**When it applies**: When routing messages to different handlers based on type.

**Copilot MUST**:
- Structure factories hierarchically: top-level → network → operation
- Use factories for operation selection
- Keep selection logic in factories, not services

**Copilot MUST NOT**:
- Use large switch statements in services
- Hardcode operation selection logic
- Mix operation selection with business logic

**Example - Correct Pattern**:

Reference: `@internal/services/authorized/operations/factory.go`

```go
// Top-level factory routes by network
type operationFactory struct {
    mastercardOperation interfaces.OperationFactory[authorizer.AuthorizedMessage]
    visaOperation       interfaces.OperationFactory[authorizer.AuthorizedMessage]
    eloOperation        interfaces.OperationFactory[authorizer.AuthorizedMessage]
}

func (f *operationFactory) GetOperation(msg *authorizer.AuthorizedMessage) (interfaces.Operation[authorizer.AuthorizedMessage], error) {
    switch msg.Product {
    case enums.PDT_MASTER, enums.PDT_MASTER_MAESTRO:
        return f.mastercardOperation.GetOperation(msg) // Delegate to network factory
    case enums.PDT_VISA, enums.PDT_VISA_ELECTRON:
        return f.visaOperation.GetOperation(msg)
    case enums.PDT_ELO_CREDIT, enums.PDT_ELO_DEBIT:
        return f.eloOperation.GetOperation(msg)
    default:
        return nil, errors.ErrOperationNotFound
    }
}
```

---

## Rule: Repository Pattern for Data Access

**Description**: Abstract all data access behind repository interfaces.

**When it applies**: When accessing databases or external data sources.

**Copilot MUST**:
- Keep repositories focused on data operations only
- Use repository interfaces for all data access
- Keep business logic out of repositories
- Use MultiDBManager for multi-tenant database access

**Copilot MUST NOT**:
- Put business logic in repositories
- Use raw database connections in services
- Mix query logic with business rules

**Example - Correct Pattern**:

Reference: `@internal/repositories/events_repository.go`

```go
// Repository only handles data access
func (pr *clearingEventsRepository) WasProcessed(
    ctx context.Context,
    customer *configuration.CustomerConfig,
    nsu, mid, authTransactionDate, authCode string,
) (bool, error) {
    ctx, span := pr.tracer.Start(ctx, "clearingEventsRepository.WasProcessed")
    defer span.End()
    
    // Data access only
    db, err := pr.multiDBManager.GetRODB(customer)
    if err != nil {
        span.RecordError(err)
        return false, errors.ErrDatabaseConnection
    }
    
    // Query logic (no business rules)
    // ...
    
    return false, nil
}
```

---

## Rule: Context Propagation

**Description**: Always propagate `context.Context` through all function calls.

**When it applies**: When creating functions that make external calls or perform I/O.

**Copilot MUST**:
- Accept `context.Context` as the first parameter
- Propagate context through all layers
- Use context for tracing, logging, and cancellation
- Create spans using context

**Copilot MUST NOT**:
- Create new contexts without reason
- Store context in structs
- Ignore context cancellation

**Example - Correct Pattern**:

Reference: `@internal/services/authorized/service.go`

```go
func (as *authorizedService) Process(
    ctx context.Context, // ✅ First parameter
    customer *configuration.CustomerConfig,
    msg *authorizer.AuthorizedMessage,
) error {
    // Propagate context to repository
    wasProcessed, err := as.eventsRepo.WasProcessed(ctx, customer, ...)
    
    // Propagate context to operation
    return op.Process(ctx, customer, msg)
}
```

---

## Rule: Error Handling and Propagation

**Description**: Always handle errors explicitly and return domain-specific errors.

**When it applies**: When handling any error condition.

**Copilot MUST**:
- Return errors, never panic in production code
- Use domain-specific error types from `go-clearing-toolkit-lib/errors`
- Wrap errors with context: `fmt.Errorf("action: %w", err)`
- Log errors before returning them

**Copilot MUST NOT**:
- Ignore errors with `_`
- Use generic `error` type when domain errors exist
- Lose error context when converting between layers
- Panic for recoverable errors

**Example - Correct Pattern**:

Reference: `@internal/transport/rmq/consumers/authorized.go`

```go
if err := json.Unmarshal(delivery.Body, &authorizedMessage); err != nil {
    logrus.WithContext(ctx).Error("invalid message, expected authorizedMessage")
    return errors.ErrFailedToUnmarshalJSON // ✅ Domain error
}

if err := authorizedMessage.Validate(); err != nil {
    logrus.WithContext(ctx).WithError(err).Errorf("invalid authorized message")
    return errors.ErrUnformattedMessage // ✅ Domain error
}
```
