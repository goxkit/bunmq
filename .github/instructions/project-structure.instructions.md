---
applyTo: "**/*"
---

# Project Structure Instructions

## Rule: Directory Structure Compliance

**Description**: Code MUST be organized according to the project's directory structure conventions.

**When it applies**: When creating new files or organizing code.

**Copilot MUST**:
- Place application entry points in `cmd/`
- Place private application code in `internal/` with proper layer separation
- Place public library code in `pkg/`
- Place generated mocks in `mocks/`
- Place test data in `data/`
- Follow the exact directory structure defined below

**Copilot MUST NOT**:
- Create files outside the defined structure
- Mix layers (e.g., business logic in transport layer)
- Place internal code in `pkg/`
- Place business logic in `cmd/`

**Directory Structure (MANDATORY)**:

```
go-clearing-mc-inc-ms/
├── cmd/                    # Application entry points
│   ├── cmd.go             # Cobra command definitions
│   └── container.go       # Dependency injection container
├── internal/               # Private application code (NOT importable)
│   ├── interfaces/        # Service and repository contracts
│   │   ├── repositories/  # Repository interfaces
│   │   └── services/      # Service interfaces
│   ├── repositories/      # Data access layer implementations
│   │   ├── chargeback_case.go
│   │   ├── fee_collection.go
│   │   ├── incoming_file_item.go
│   │   ├── incoming_file.go
│   │   └── transaction.go
│   ├── services/          # Business logic layer
│   │   ├── incoming/      # T112 incoming file processing
│   │   │   ├── svc.go
│   │   │   ├── file_mover.go
│   │   │   └── operations/ # Record type operations
│   │   ├── full_mpe/      # Full MPE processing
│   │   ├── incremental_mpe/ # Incremental MPE processing
│   │   ├── files_management/ # Agnostic file operations
│   │   └── ...            # Other service domains
│   └── transport/         # External communication adapters
│       ├── http/          # HTTP handlers and routes
│       │   ├── handlers/
│       │   ├── routes/
│       │   └── viewmodels/
│       └── rmq/           # RabbitMQ consumers
│           └── consumers/
├── pkg/                   # Public library code (importable)
│   └── timezone/          # Timezone utilities
├── mocks/                 # Generated mocks for testing
├── data/                  # Test data files (JSON fixtures)
├── main.go                # Application entry point
├── go.mod                 # Go module definition
├── Dockerfile             # Container definition
└── Makefile               # Build automation
```

---

## Rule: Package Naming

**Description**: Package names MUST follow Go conventions and match directory names.

**When it applies**: When creating new packages or files.

**Copilot MUST**:
- Use lowercase, single-word package names
- Match package name exactly to directory name
- Use descriptive names that indicate package purpose
- Keep package names concise

**Copilot MUST NOT**:
- Use underscores or mixedCaps in package names
- Use abbreviations without context
- Create package names that don't match directory names

**Example - Correct Package Naming**:

```go
// ✅ CORRECT: Package name matches directory
// File: internal/services/authorized/service.go
package authorized

// ✅ CORRECT: Package name matches directory
// File: internal/repositories/trx_repository.go
package repositories
```

**Example - Incorrect Package Naming**:

```go
// ❌ WRONG: Package name doesn't match directory
// File: internal/services/authorized/service.go
package authorizedService

// ❌ WRONG: Underscore in package name
// File: internal/repositories/trx_repository.go
package trx_repository
```

---

## Rule: File Naming

**Description**: File names MUST use snake_case and match primary type/function.

**When it applies**: When creating new files.

**Copilot MUST**:
- Use snake_case for file names (lowercase with underscores)
- Match file name to primary exported type or function
- Use `_test.go` suffix for test files
- Keep file names descriptive

**Copilot MUST NOT**:
- Use camelCase or PascalCase for file names
- Create files with generic names (`utils.go`, `helpers.go`)
- Use abbreviations in file names

**Example - Correct File Naming**:

```
✅ internal/services/authorized/service.go
✅ internal/repositories/trx_repository.go
✅ internal/services/authorized/operations/visa/credit.go
✅ internal/services/authorized/operations/visa/credit_test.go
```

**Example - Incorrect File Naming**:

```
❌ internal/services/authorized/authorizedService.go
❌ internal/repositories/TrxRepository.go
❌ internal/services/authorized/operations/visa/creditTest.go
```

---

## Rule: Layer Separation

**Description**: Code MUST be placed in the correct architectural layer.

**When it applies**: When creating or moving code.

**Copilot MUST**:
- Place transport/adapter code in `internal/transport/`
- Place business logic in `internal/services/`
- Place data access in `internal/repositories/`
- Place domain entities in `internal/models/`
- Place interfaces in `internal/interfaces/`

**Copilot MUST NOT**:
- Place business logic in transport layer
- Place data access in service layer
- Mix concerns across layers
- Create circular dependencies between layers

**Example - Correct Layer Placement**:

Reference: `@internal/transport/rmq/consumers/authorized.go`

```go
// ✅ CORRECT: Transport layer - orchestration only
package consumers

func (c *authorizedConsumer) Handle(ctx context.Context, delivery *amqp091.Delivery) error {
    // Unmarshal, validate, delegate to service
    return c.authorizerService.Process(ctx, customer, &authorizedMessage)
}
```

Reference: `@internal/services/authorized/service.go`

```go
// ✅ CORRECT: Service layer - business logic
package authorized

func (as *authorizedService) Process(ctx context.Context, ...) error {
    // Business logic here
}
```

---

## Rule: Module Path

**Description**: All internal imports MUST use the module path prefix.

**When it applies**: When generating import statements.

**Copilot MUST**:
- Use module path: `bitbucket.org/asappay/go-clearing-mc-inc-ms`
- Import internal packages using full module path
- Organize imports: standard library, third-party, internal

**Copilot MUST NOT**:
- Use relative imports
- Use shortened import paths
- Mix import styles

**Example - Correct Import Paths**:

Reference: `@internal/services/incoming/svc.go`

```go
import (
    "context"
    "database/sql"
    
    "bitbucket.org/asappay/go-clearing-models-lib/messages"
    "github.com/sirupsen/logrus"
    
    "bitbucket.org/asappay/go-clearing-mc-inc-ms/internal/interfaces/repositories"
    "bitbucket.org/asappay/go-clearing-mc-inc-ms/internal/interfaces/services"
    "bitbucket.org/asappay/go-clearing-toolkit-lib/configs"
)
```

---

## Rule: Test File Organization

**Description**: Test files MUST be placed alongside source files in the same package.

**When it applies**: When creating test files.

**Copilot MUST**:
- Place test files in the same directory as source files
- Use `_test.go` suffix
- Keep test files in the same package (unless testing exported API)
- Organize tests to match source code structure

**Copilot MUST NOT**:
- Create separate test directories
- Place tests in different packages unnecessarily
- Mix test files with source files inappropriately

**Example - Correct Test Organization**:

```
internal/services/authorized/
├── service.go          # Source file
└── service_test.go     # ✅ Test file in same directory
```

---

## Rule: Configuration Files

**Description**: Configuration MUST be loaded from environment variables, never hardcoded.

**When it applies**: When generating configuration or initialization code.

**Copilot MUST**:
- Load configuration from environment variables
- Use `configs.LoadEnvironmentProps()` for environment loading
- Use `ConfigTool` for remote configuration
- Document required environment variables

**Copilot MUST NOT**:
- Hardcode configuration values
- Store secrets in configuration files
- Use configuration files committed to version control for secrets

**Example - Correct Configuration**:

Reference: `@cmd/container.go`

```go
// ✅ CORRECT: Configuration from environment
configs.LoadEnvironmentProps()
logger.Init(
    configs.Config.GetString("ENVIRONMENT"),
    configs.Config.GetString("LOG_LEVEL"),
)

// ✅ CORRECT: Remote configuration for sensitive data
configTool, err := configTool.NewConfigTool()
```

---

## Project Structure Checklist

When generating code, ensure:

- ✅ Files placed in correct directories according to layer
- ✅ Package names match directory names
- ✅ File names use snake_case
- ✅ Imports use full module path
- ✅ Test files placed alongside source files
- ✅ Configuration loaded from environment variables
- ✅ No circular dependencies
- ✅ Clear layer separation maintained
