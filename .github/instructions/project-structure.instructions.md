---
applyTo: "**/*"
---

# Project Structure Instructions

## Rule: Directory Structure Compliance

**Description**: Code MUST be organized according to the project's directory structure conventions.

**When it applies**: When creating new files or organizing code.

**Copilot MUST**:
- Place all library code in the root directory (flat structure)
- Place example code in `examples/`
- Place test files alongside source files with `_test.go` suffix
- Follow the exact directory structure defined below

**Copilot MUST NOT**:
- Create unnecessary subdirectories
- Create `internal/` or `pkg/` directories (this is a library, not an application)
- Mix test files with non-test concerns
- Create separate test directories

**Directory Structure (MANDATORY)**:

```
bunmq/
├── connection.go           # RMQConnection interface definition
├── connection_manager.go   # Connection management with auto-reconnection
├── channel.go              # Channel wrapper and management
├── topology.go             # Topology builder for exchanges, queues, bindings
├── queue.go                # Queue definitions with retry/DLQ configuration
├── exchange.go             # Exchange definitions (direct, topic, fanout, headers)
├── binding.go              # Binding configuration between exchanges and queues
├── publisher.go            # Message publishing with tracing and deadlines
├── dispatcher.go           # Message consumption and handler dispatching
├── tracing.go              # OpenTelemetry integration for AMQP headers
├── options.go              # Publishing options and message metadata
├── errors.go               # Domain-specific error types
├── mocks.go                # Mock implementations for testing
├── *_test.go               # Unit tests alongside source files
├── examples/               # Usage examples
│   └── main.go             # Example application demonstrating library usage
├── .github/                # GitHub-specific configurations
│   ├── copilot-instructions.md    # Main Copilot instruction file
│   ├── instructions/       # Path-specific instruction files
│   ├── workflows/          # CI/CD workflows
│   └── CODEOWNERS          # Code ownership configuration
├── Makefile                # Build automation (test, lint, sec)
├── .gosec.json             # Security scanner configuration
├── codecov.yml             # Code coverage configuration
├── go.mod                  # Module definition (github.com/goxkit/bunmq)
└── go.sum                  # Dependency checksums
```

---

## Rule: Package Naming

**Description**: Package name MUST be `bunmq` for all library files.

**When it applies**: When creating new packages or files.

**Copilot MUST**:
- Use `package bunmq` for all library source files
- Use `package main` only for example applications
- Keep package name consistent across all library files

**Copilot MUST NOT**:
- Use different package names for library files
- Create subpackages unless absolutely necessary
- Use underscores or mixedCaps in package names

**Example - Correct Package Naming**:

Reference: `@connection.go`, `@publisher.go`, `@dispatcher.go`

```go
// ✅ CORRECT: All library files use the same package name
// File: connection.go
package bunmq

// File: publisher.go
package bunmq

// File: dispatcher.go
package bunmq
```

**Example - Correct Example Package**:

Reference: `@examples/main.go`

```go
// ✅ CORRECT: Example application uses package main
package main

import (
    "github.com/goxkit/bunmq"
)
```

---

## Rule: File Naming

**Description**: File names MUST use snake_case and describe the primary concept.

**When it applies**: When creating new files.

**Copilot MUST**:
- Use snake_case for file names (lowercase with underscores)
- Name files after the primary concept they implement
- Use `_test.go` suffix for test files
- Keep file names descriptive and concise

**Copilot MUST NOT**:
- Use camelCase or PascalCase for file names
- Create files with generic names (`utils.go`, `helpers.go`, `common.go`)
- Use abbreviations that aren't universally understood

**Example - Correct File Naming**:

```
✅ connection.go          # Connection interface and types
✅ connection_manager.go  # Connection manager implementation
✅ connection_test.go     # Tests for connection.go
✅ connection_manager_test.go  # Tests for connection_manager.go
✅ publisher.go           # Publisher interface and implementation
✅ publisher_test.go      # Tests for publisher.go
```

**Example - Incorrect File Naming**:

```
❌ Connection.go           # PascalCase not allowed
❌ connectionManager.go    # camelCase not allowed
❌ conn.go                 # Abbreviation not clear
❌ utils.go                # Generic name not allowed
❌ helpers.go              # Generic name not allowed
```

---

## Rule: Module Path and Imports

**Description**: All imports MUST use the correct module path.

**When it applies**: When generating import statements.

**Copilot MUST**:
- Use module path: `github.com/goxkit/bunmq`
- Organize imports in groups: standard library, third-party, internal
- Separate import groups with blank lines

**Copilot MUST NOT**:
- Use relative imports
- Use incorrect module paths
- Mix import groups without separation

**Example - Correct Import Organization**:

Reference: `@publisher.go`

```go
import (
    // Standard library
    "context"
    "encoding/json"
    "fmt"
    "reflect"
    "time"

    // Third-party dependencies
    "github.com/google/uuid"
    amqp "github.com/rabbitmq/amqp091-go"
    "github.com/sirupsen/logrus"
)
```

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

## Rule: Test File Organization

**Description**: Test files MUST be placed alongside source files.

**When it applies**: When creating test files.

**Copilot MUST**:
- Place test files in the same directory as source files
- Use `_test.go` suffix
- Use `package bunmq` for internal tests (access to private members)
- Name test functions as `Test<FunctionName>_<Scenario>`

**Copilot MUST NOT**:
- Create separate test directories
- Use `package bunmq_test` unless testing only exported API
- Mix test files with example files

**Example - Correct Test Organization**:

```
bunmq/
├── connection.go          # Source file
├── connection_test.go     # ✅ Test file in same directory
├── publisher.go           # Source file
├── publisher_test.go      # ✅ Test file in same directory
├── dispatcher.go          # Source file
└── dispatcher_test.go     # ✅ Test file in same directory
```

---

## Rule: Configuration Files

**Description**: Configuration files MUST be placed in the root directory with appropriate naming.

**When it applies**: When creating or modifying configuration files.

**Copilot MUST**:
- Place Makefile in root directory
- Place .gosec.json for security configuration
- Place codecov.yml for coverage configuration
- Place go.mod and go.sum in root directory
- Place GitHub configurations in .github/

**Copilot MUST NOT**:
- Create configuration files in subdirectories (except .github/)
- Use non-standard configuration file names
- Duplicate configuration files

---

## Project Structure Checklist

When generating code, ensure:

- ✅ File placed in correct location (root for library, examples/ for examples)
- ✅ Package name is `bunmq` for library files
- ✅ File names use snake_case
- ✅ Imports use correct module path (github.com/goxkit/bunmq)
- ✅ Test files placed alongside source files
- ✅ No unnecessary subdirectories created
- ✅ Import groups properly organized and separated
