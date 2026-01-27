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
- Use short names for common types: `ctx`, `db`, `err`, `msg`

**Copilot MUST NOT**:
- Use single-letter names (except `i`, `j`, `k` for iterators)
- Use abbreviations unless universally understood
- Use generic names (`data`, `info`, `temp`, `obj`)
- Mix naming conventions

**Example - Correct Naming**:

Reference: `@internal/services/authorized/service.go`

```go
// ✅ CORRECT: Exported type with PascalCase
type AuthorizedService interface {
    Process(ctx context.Context, ...) error
}

// ✅ CORRECT: Unexported struct with camelCase
type authorizedService struct {
    operationFactory interfaces.OperationFactory[authorizer.AuthorizedMessage]
    eventsRepo       interfaces.EventsRepository
}

// ✅ CORRECT: Constructor with New prefix
func NewAuthorizedService(...) interfaces.AuthorizerService {
    return &authorizedService{...}
}

// ✅ CORRECT: Method with verb name
func (as *authorizedService) Process(ctx context.Context, ...) error {
    // ...
}
```

**Example - Incorrect Naming**:

```go
// ❌ WRONG: Single-letter variable for important data
func Process(m *authorizer.AuthorizedMessage) error {
    // ...
}

// ❌ WRONG: Generic name
func Process(data interface{}) error {
    // ...
}

// ❌ WRONG: Abbreviation without context
func ProcAuthMsg(msg *authorizer.AuthorizedMessage) error {
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
- Use early returns to reduce nesting
- Ensure single responsibility per function

**Copilot MUST NOT**:
- Create functions over 50 lines
- Allow deep nesting (> 3 levels)
- Mix multiple concerns in one function
- Create god functions

**Example - Correct Function Size**:

Reference: `@internal/services/authorized/service.go`

```go
// ✅ CORRECT: Small, focused function
func (as *authorizedService) Process(ctx context.Context, customer *configuration.CustomerConfig, msg *authorizer.AuthorizedMessage) error {
    if msg == nil {
        logrus.WithContext(ctx).Error("authorizedMessage cannot be nil")
        return errors.ErrUnformattedMessage
    }
    
    db, err := as.uow.GetW(customer)
    if err != nil {
        logrus.WithContext(ctx).WithError(err).Error("failed to get db connection")
        return errors.ErrDatabaseConnection
    }
    
    if err := as.uow.Begin(ctx, db); err != nil {
        logrus.WithContext(ctx).WithError(err).Error("failed to begin transaction")
        return errors.ErrDatabaseConnection
    }
    
    defer func() {
        _ = as.uow.Rollback(ctx, db)
    }()
    
    // Business logic delegated to helper methods
    return as.processMessage(ctx, db, customer, msg)
}
```

---

## Rule: Code Formatting

**Description**: Code MUST be formatted according to `gofmt` and `goimports` standards.

**When it applies**: When generating any Go code.

**Copilot MUST**:
- Format code according to `gofmt` standards
- Organize imports with `goimports` (stdlib, third-party, internal)
- Use tabs for indentation
- Keep lines under 120 characters when possible

**Copilot MUST NOT**:
- Use non-standard formatting
- Mix tabs and spaces
- Create lines exceeding 120 characters unnecessarily

**Example - Correct Import Organization**:

Reference: `@internal/services/authorized/service.go`

```go
import (
    // Standard library
    "context"
    
    // Third-party
    "bitbucket.org/asappay/go-acquirer-messages-lib/authorizer"
    "bitbucket.org/asapshared/go-tools/v2/config_tool/configuration"
    "github.com/sirupsen/logrus"
    
    // Internal
    "bitbucket.org/asappay/go-clearing-events-ms/internal/services/interfaces"
    "bitbucket.org/asappay/go-clearing-toolkit-lib/errors"
)
```

---

## Rule: Comments and Documentation

**Description**: All exported symbols MUST have documentation comments.

**When it applies**: When creating exported types, functions, or packages.

**Copilot MUST**:
- Add package comments explaining package purpose
- Document all exported functions with clear descriptions
- Document all exported types
- Explain "why" in comments, not "what"
- Start comments with the symbol name

**Copilot MUST NOT**:
- Skip documentation for exported symbols
- Add obvious comments that restate code
- Document implementation details unnecessarily

**Example - Correct Documentation**:

```go
// Package authorized provides services for processing authorized transactions
// from payment networks (Visa, Mastercard, Elo).
package authorized

// AuthorizedService processes authorized transaction messages from payment networks.
// It orchestrates the conversion of authorized messages to clearing transactions.
type AuthorizedService interface {
    // Process handles an authorized message and creates clearing transactions.
    // It returns an error if processing fails.
    Process(ctx context.Context, customer *configuration.CustomerConfig, msg *authorizer.AuthorizedMessage) error
}

// NewAuthorizedService creates a new authorized service with the given dependencies.
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

---

## Rule: Magic Numbers and Constants

**Description**: NEVER use magic numbers. Always use named constants.

**When it applies**: When using numeric literals or string literals for configuration.

**Copilot MUST**:
- Define constants for all numeric literals (except 0, 1, -1)
- Use descriptive constant names
- Group related constants together
- Use constants for configuration keys

**Copilot MUST NOT**:
- Use magic numbers in code
- Use string literals for configuration keys repeatedly
- Hardcode values that should be configurable

**Example - Correct Constant Usage**:

```go
const (
    defaultTimeout     = 30 * time.Second
    maxRetries         = 3
    queueMaxLength     = 100000
)

// ✅ CORRECT: Using named constant
time.Sleep(defaultTimeout)

// ❌ WRONG: Magic number
time.Sleep(30 * time.Second)
```

---

## Rule: Error Messages

**Description**: Error messages MUST be descriptive and actionable.

**When it applies**: When creating error messages or log messages.

**Copilot MUST**:
- Explain what went wrong
- Include relevant context (IDs, values when safe)
- Use consistent format
- Be user-friendly when exposed to users

**Copilot MUST NOT**:
- Use generic messages ("error occurred")
- Include sensitive information
- Use technical jargon for user-facing errors

**Example - Correct Error Messages**:

Reference: `@internal/services/authorized/service.go`

```go
// ✅ CORRECT: Descriptive error message with context
logrus.WithContext(ctx).
    WithError(err).
    WithFields(msg.LogFields()).
    Error("failed to insert authorized event in database")

// ❌ WRONG: Generic error message
logrus.Error("error")
```

---

## Rule: Struct Organization

**Description**: Struct fields MUST be organized logically.

**When it applies**: When creating structs.

**Copilot MUST**:
- Organize fields: exported first, then unexported
- Group related fields together
- Use field tags for JSON, database, validation
- Provide constructors for structs that need initialization

**Example - Correct Struct Organization**:

Reference: `@internal/services/authorized/service.go`

```go
// ✅ CORRECT: Exported fields first, then unexported
type authorizedService struct {
    operationFactory interfaces.OperationFactory[authorizer.AuthorizedMessage]
    eventsRepo       interfaces.EventsRepository
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
- Use `var` when short declaration is possible
- Create variables with unclear purpose

**Example - Correct Variable Declarations**:

```go
// ✅ CORRECT: Short declaration close to use
db, err := as.uow.GetW(customer)
if err != nil {
    return errors.ErrDatabaseConnection
}

// ✅ CORRECT: Meaningful names
authorizedMessage := &authorizer.AuthorizedMessage{}

// ❌ WRONG: Generic name
msg := &authorizer.AuthorizedMessage{}
```

---

## Coding Conventions Checklist

When generating code, ensure:

- ✅ PascalCase for exported, camelCase for unexported
- ✅ Descriptive, intention-revealing names
- ✅ Functions under 30 lines
- ✅ Proper import organization
- ✅ All exported symbols documented
- ✅ No magic numbers (use constants)
- ✅ Descriptive error messages
- ✅ Code formatted with gofmt/goimports
- ✅ Early returns to reduce nesting
- ✅ Single responsibility per function
