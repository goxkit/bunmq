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

**Tool Reference**: See `@Makefile` lines 79-81 for lint configuration.

---

## Rule: Format Code with gofmt

**Description**: All code MUST be formatted according to `gofmt` standards.

**When it applies**: When generating any Go code.

**Copilot MUST**:
- Format code according to `gofmt` standards
- Organize imports with `goimports`
- Use 4-space indentation (tabs converted to spaces)
- Keep lines under 100 characters

**Copilot MUST NOT**:
- Use non-standard formatting
- Mix formatting styles
- Generate lines exceeding 100 characters

---

## Rule: Enforce gosec for Security

**Description**: All dependencies MUST pass `gosec` security checks.

**When it applies**: When adding dependencies or generating code that uses external packages.

**Copilot MUST**:
- Use specific version constraints in `go.mod` (avoid `*` wildcards)
- Prefer well-maintained packages
- Check for known CVEs before adding dependencies

**Copilot MUST NOT**:
- Add dependencies with known CVEs
- Use `*` version constraints
- Ignore security warnings

**Example - Correct Dependency Declaration**:

Reference: `@go.mod`

```go
require (
	bitbucket.org/asapcard/go-acquirer-enums-core-lib v0.11.1 // ✅ Specific version
	github.com/go-chi/chi/v5 v5.2.3                          // ✅ Specific version
)
```

---

## Rule: Error Handling

**Description**: Always handle errors explicitly using domain-specific error types.

**When it applies**: When handling any error condition.

**Copilot MUST**:
- Use domain-specific error types from `go-clearing-toolkit-lib/errors`
- Wrap errors with context: `fmt.Errorf("action: %w", err)`
- Use `errors.Is()` and `errors.As()` for error checking
- Log errors before returning them

**Copilot MUST NOT**:
- Use `panic()` for recoverable errors
- Return string errors
- Ignore errors with `_`
- Lose error context

**Example - Correct Error Handling**:

Reference: `@internal/transport/rmq/consumers/authorized.go`

```go
if err := json.Unmarshal(delivery.Body, &authorizedMessage); err != nil {
    logrus.WithContext(ctx).Error("invalid message, expected authorizedMessage")
    return errors.ErrFailedToUnmarshalJSON // ✅ Domain error
}
```

---

## Rule: Context Management

**Description**: Always accept `context.Context` as the first parameter for I/O operations.

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
    // Propagate context
    wasProcessed, err := as.eventsRepo.WasProcessed(ctx, customer, ...)
    // ...
}
```

---

## Rule: Interface Design

**Description**: Define interfaces near consumers, keep them small and focused.

**When it applies**: When creating new interfaces or dependencies.

**Copilot MUST**:
- Define interfaces in `internal/interfaces/` or near consumers
- Keep interfaces small and focused (Interface Segregation Principle)
- Use interfaces for dependency injection
- Use generic interfaces when appropriate

**Copilot MUST NOT**:
- Create interfaces with single implementations (unless for testing)
- Put interface definitions in implementation files
- Create large interfaces that violate ISP

**Example - Correct Interface Design**:

Reference: `@internal/services/interfaces/operation.go`

```go
// Generic interface for operations
type Operation[T any] interface {
    Process(ctx context.Context, customer *configuration.CustomerConfig, msg *T) error
}

type OperationFactory[T any] interface {
    GetOperation(msg *T) (Operation[T], error)
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
- Use field tags for JSON, database, validation

**Example - Correct Struct Organization**:

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

---

## Rule: Naming Conventions

**Description**: Follow Go naming conventions strictly.

**When it applies**: When naming any identifier.

**Copilot MUST**:
- Use PascalCase for exported names
- Use camelCase for unexported names
- Use descriptive, intention-revealing names
- Follow domain terminology consistently

**Copilot MUST NOT**:
- Use single-letter names (except `i`, `j`, `k` for iterators)
- Use abbreviations unless universally understood
- Use generic names (`data`, `info`, `temp`)

**Example - Correct Naming**:

Reference: `@internal/services/authorized/service.go`

```go
type authorizedService struct { // ✅ Descriptive, unexported
    operationFactory interfaces.OperationFactory[authorizer.AuthorizedMessage] // ✅ Clear purpose
    eventsRepo       interfaces.EventsRepository // ✅ Clear purpose
}

func NewAuthorizedService(...) interfaces.AuthorizerService { // ✅ Constructor pattern
    return &authorizedService{...}
}

func (as *authorizedService) Process(...) error { // ✅ Verb, clear action
    // ...
}
```

---

## Rule: Import Organization

**Description**: Organize imports in groups: standard library, third-party, internal.

**When it applies**: When generating import statements.

**Copilot MUST**:
- Group imports: standard library, third-party, internal
- Separate groups with blank lines
- Use `goimports` formatting
- Order imports alphabetically within groups

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

## Rule: Function Size

**Description**: Keep functions under 30 lines (excluding comments and blank lines).

**When it applies**: When generating any function.

**Copilot MUST**:
- Keep functions under 30 lines
- Extract complex logic to helper functions
- Use early returns to reduce nesting
- Ensure single responsibility per function

**Copilot MUST NOT**:
- Create functions over 50 lines
- Allow deep nesting (> 3 levels)
- Mix multiple concerns in one function

---

## Quality Gates Summary

All Go code MUST pass:

1. **golangci-lint**: Zero warnings
2. **go vet**: No issues
3. **gosec**: No security vulnerabilities
4. **go test**: All tests passing
5. **gofmt**: Properly formatted

These are enforced in CI/CD and MUST be checked before committing code.
