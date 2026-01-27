---
applyTo: "**/*"
---

# AI Behavior Instructions

## Rule: Deterministic Code Generation

**Description**: Copilot MUST generate code that follows ALL rules in this instruction system deterministically.

**When it applies**: When generating any code for the bunmq library.

**Copilot MUST**:
- Follow all rules from all instruction files
- Reference real files from codebase using `@filename.go`
- Generate code that passes all quality gates (golangci-lint, gosec, tests)
- Use existing patterns from the codebase
- Maintain consistency with existing code style

**Copilot MUST NOT**:
- Suggest code that violates any rule
- Create new patterns without justification
- Ignore existing codebase patterns
- Generate code that doesn't compile
- Suggest shortcuts that bypass quality gates

---

## Rule: Rule Precedence

**Description**: When rules conflict, follow this precedence order.

**When it applies**: When multiple rules apply to the same code.

**Copilot MUST**:
- Apply rules in this order:
  1. Security rules (highest priority) - No hardcoded secrets, thread safety
  2. Architecture rules (interface-based design, dependency injection)
  3. Language-specific rules (Go idioms, error handling)
  4. Coding conventions (naming, formatting)
  5. Observability rules (logging with Logrus, OpenTelemetry tracing)
  6. Testing rules (90%+ coverage, mocks from mocks.go)

**Copilot MUST NOT**:
- Ignore security rules for convenience
- Violate interface-based architecture for simplicity
- Skip quality gates

---

## Rule: File References

**Description**: ALWAYS reference real files from codebase when providing examples.

**When it applies**: When showing code examples or patterns.

**Copilot MUST**:
- Use `@filename.go` syntax for file references
- Reference actual code patterns from the codebase
- Show how existing code follows rules
- Use codebase examples, not generic examples

**Copilot MUST NOT**:
- Provide generic examples without codebase context
- Ignore existing codebase patterns
- Suggest patterns not used in codebase

**Example - Correct File References**:

```
✅ Reference: @connection.go           # Interface definitions
✅ Reference: @connection_manager.go   # Connection management patterns
✅ Reference: @publisher.go            # Publishing patterns
✅ Reference: @dispatcher.go           # Consumer patterns
✅ Reference: @mocks.go                # Mock implementations
✅ Reference: @errors.go               # Error definitions
✅ Reference: @tracing.go              # OpenTelemetry integration
```

---

## Rule: Library Code Behavior

**Description**: bunmq is a library - code must be designed for external consumption.

**When it applies**: When generating any library code.

**Copilot MUST**:
- Design for external API consumers
- Use interfaces for all public components
- Never panic - always return errors
- Support graceful shutdown
- Provide thread-safe implementations
- Include comprehensive documentation

**Copilot MUST NOT**:
- Panic in library code (except init for programmer errors)
- Expose internal implementation details
- Create breaking API changes
- Hard-code configuration values
- Log connection strings or sensitive data

**Example - Library Pattern**:

Reference: `@publisher.go`

```go
// ✅ Interface for external consumption
type Publisher interface {
    Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error
    PublishDeadline(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error
}

// ✅ Constructor returns interface, not concrete type
func NewPublisher(appName string, connManager ConnectionManager) Publisher {
    return &publisher{
        appName:     appName,
        connManager: connManager,
    }
}
```

---

## Rule: Code Modification Behavior

**Description**: When modifying existing code, preserve existing patterns and architecture.

**When it applies**: When editing or refactoring existing code.

**Copilot MUST**:
- Maintain existing interface-based architecture
- Follow existing naming conventions (PascalCase for exported, camelCase for private)
- Preserve existing error handling patterns (wrap with context)
- Keep existing logging patterns (Logrus with structured fields)
- Maintain existing test structure (table-driven tests)

**Copilot MUST NOT**:
- Refactor without explicit request
- Change interface signatures without discussion
- Modify working code unnecessarily
- Break existing tests
- Change patterns just for "improvement"

---

## Rule: New Code Creation

**Description**: When creating new code, follow all rules and match existing patterns.

**When it applies**: When generating new files or functions.

**Copilot MUST**:
- Follow interface-based design
- Use dependency injection via constructor parameters
- Implement interfaces from `connection.go`, `channel.go`
- Add structured logging with Logrus
- Add OpenTelemetry tracing for significant operations
- Write tests using mocks from `mocks.go`
- Follow naming conventions

**Copilot MUST NOT**:
- Create code that violates interface-based architecture
- Skip dependency injection
- Ignore logging requirements
- Skip tests
- Create new patterns without justification

---

## Rule: Code Review Before Suggesting

**Description**: Verify code against ALL applicable rules before suggesting.

**When it applies**: Before suggesting any code.

**Copilot MUST CHECK**:
- Does it follow interface-based design?
- Does it use dependency injection?
- Does it handle errors properly (no ignored errors)?
- Does it include structured logging with Logrus?
- Does it include OpenTelemetry tracing?
- Does it follow naming conventions?
- Does it pass quality gates (golangci-lint, gosec)?
- Does it have tests with 90%+ coverage?
- Does it reference existing patterns from the codebase?

---

## Rule: Complete Code Suggestions

**Description**: Provide complete, working code that follows all rules.

**When it applies**: When suggesting code implementations.

**Copilot MUST**:
- Provide complete implementations
- Include error handling with wrapping
- Include logging with context (`logrus.WithContext(ctx)`)
- Include necessary imports
- Follow all coding conventions
- Reference real codebase patterns
- Include tests when appropriate

**Copilot MUST NOT**:
- Suggest incomplete code
- Skip error handling
- Skip logging
- Suggest code that violates rules
- Use patterns not in codebase
- Leave TODOs without implementation

**Example - Complete Implementation**:

Reference: `@publisher.go`

```go
// ✅ Complete function with error handling and logging
func (p *publisher) Publish(ctx context.Context, exchange, routingKey string, msg any, options ...*Option) error {
    // ✅ Input validation
    if exchange == "" {
        logrus.WithContext(ctx).Error("exchange cannot be empty")
        return fmt.Errorf("exchange cannot be empty")
    }

    // ✅ Delegation with error handling
    return p.publish(ctx, exchange, routingKey, msg, options...)
}
```

---

## Rule: Question Answering

**Description**: When answering questions, reference actual codebase files and patterns.

**When it applies**: When explaining code or architecture.

**Copilot MUST**:
- Reference real files using `@filename.go`
- Explain how bunmq implements patterns
- Show actual code examples from the codebase
- Reference relevant instruction files
- Provide bunmq-specific context

**Copilot MUST NOT**:
- Provide generic answers without codebase context
- Ignore existing codebase patterns
- Suggest patterns not used in codebase
- Give theoretical answers when codebase examples exist

---

## Rule: Error Prevention

**Description**: Actively prevent common mistakes in RabbitMQ library code.

**When it applies**: When generating any code.

**Copilot MUST AVOID**:
- Ignoring errors (always handle or wrap)
- Hardcoding connection strings
- Missing context in logs
- Skipping input validation
- Missing error handling
- Forgetting to close channels/connections
- Missing tests
- Unprotected concurrent access (use sync.Mutex)
- Leaking goroutines (use context cancellation)
- Panicking in library code

---

## Rule: Proactive Suggestions

**Description**: Point out rule violations and suggest fixes.

**When it applies**: When detecting code that violates rules.

**Copilot MUST**:
- Point out rule violations
- Suggest fixes for violations
- Explain why rules exist
- Reference relevant instruction files
- Provide corrected code examples

---

## Rule: Instruction System Awareness

**Description**: Be aware of all instruction files and their purposes.

**When it applies**: When generating code or answering questions.

**Copilot MUST KNOW**:
- `.github/copilot-instructions.md`: Main entry point, global rules, bunmq overview
- `.github/instructions/project-structure.instructions.md`: Flat library structure
- `.github/instructions/architecture.instructions.md`: Interface-based design, SOLID
- `.github/instructions/golang.instructions.md`: Go-specific rules, tooling
- `.github/instructions/coding-conventions.instructions.md`: Naming, formatting
- `.github/instructions/observability.instructions.md`: Logrus logging, OpenTelemetry tracing
- `.github/instructions/security.instructions.md`: No hardcoded secrets, thread safety
- `.github/instructions/testing.instructions.md`: Testing patterns, mocks.go usage
- `.github/instructions/ai-behavior.instructions.md`: This file

---

## Rule: Quality Gate Enforcement

**Description**: Ensure all generated code passes quality gates.

**When it applies**: When generating any code.

**Copilot MUST VERIFY**:
- Code compiles (`go build ./...`)
- Passes golangci-lint (`make lint`)
- Passes gosec (`make sec`)
- Has tests with 90%+ coverage (`make test-coverage-threshold`)
- Tests pass (`make test`)
- Follows formatting rules (`gofmt`)

**Copilot MUST NOT**:
- Suggest code that doesn't compile
- Ignore lint warnings
- Skip security checks
- Suggest untested code for public functions
- Suggest code that fails quality gates

---

## Rule: Documentation Generation

**Description**: Generate appropriate documentation comments.

**When it applies**: When creating exported symbols.

**Copilot MUST**:
- Add package comments (package bunmq provides...)
- Document exported interfaces
- Document exported functions
- Document exported types
- Explain complex logic

**Copilot MUST NOT**:
- Skip documentation for exported symbols
- Add obvious comments
- Document implementation details unnecessarily

**Example - Correct Documentation**:

Reference: `@connection_manager.go`

```go
// ConnectionManager manages RabbitMQ connections and channels with automatic reconnection.
// It monitors channel and connection health using NotifyClose and NotifyCancel.
type ConnectionManager interface {
    // GetConnection returns the current connection, ensuring it's healthy.
    GetConnection() (RMQConnection, error)

    // GetChannel returns the current channel, ensuring it's healthy.
    GetChannel() (AMQPChannel, error)

    // Close gracefully closes the connection manager.
    Close() error

    // IsHealthy checks if both connection and channel are healthy.
    IsHealthy() bool
}
```

---

## Rule: Refactoring Behavior

**Description**: Preserve functionality and follow all rules when refactoring.

**When it applies**: When refactoring code.

**Copilot MUST**:
- Maintain existing functionality
- Follow all rules
- Update tests if needed
- Preserve interface-based architecture
- Update documentation

**Copilot MUST NOT**:
- Break existing functionality
- Violate rules during refactoring
- Skip test updates
- Change public interfaces unnecessarily

---

## AI Behavior Checklist

When generating bunmq code, ensure:

- ✅ All rules followed deterministically
- ✅ Real codebase files referenced (`@filename.go`)
- ✅ Interface-based design maintained
- ✅ Quality gates passed (lint, sec, test)
- ✅ Complete implementations provided
- ✅ Error handling with wrapping included
- ✅ Structured logging with Logrus included
- ✅ OpenTelemetry tracing for significant operations
- ✅ Tests using mocks from mocks.go
- ✅ Documentation for exported symbols
- ✅ Thread-safe code (sync.Mutex where needed)
- ✅ No panics in library code

---

## Anti-Patterns for AI (FORBIDDEN)

**Copilot MUST NEVER**:
1. Suggest code that violates any rule
2. Ignore existing codebase patterns
3. Create new patterns without justification
4. Skip error handling
5. Skip structured logging
6. Skip tests for public functions
7. Suggest code that doesn't compile
8. Ignore security rules (hardcoded secrets, etc.)
9. Break interface-based architecture
10. Provide generic examples without bunmq context
11. Use panic in library code
12. Ignore thread safety requirements
13. Leave resources uncleaned (goroutines, channels)
14. Log connection strings or sensitive data

---

## Positive AI Behaviors (REQUIRED)

**Copilot MUST ALWAYS**:
1. Reference real codebase files (`@publisher.go`, `@dispatcher.go`, etc.)
2. Follow all rules deterministically
3. Provide complete, working code
4. Include error handling with wrapping
5. Include logging with `logrus.WithContext(ctx)`
6. Include tests using mocks from `mocks.go`
7. Explain rule application
8. Verify quality gates
9. Use existing patterns from bunmq
10. Maintain interface-based architecture
11. Ensure thread safety with sync.Mutex
12. Support graceful shutdown with context
13. Use OpenTelemetry for tracing
14. Document exported symbols
