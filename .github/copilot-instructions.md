# bunmq - RabbitMQ Go Library

bunmq is a Go library that provides high-level abstractions for RabbitMQ messaging with automatic connection recovery, retry mechanisms, and dead letter queue support. This is a **library** package, not an executable application.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap and Build
```bash
# Install Go dependencies (takes ~2 seconds)
go mod download

# Build all packages (takes <1 second)
go build ./...

# Run tests (takes ~14 seconds) - NEVER CANCEL
go test ./...
```

### Linting and Security Checks
```bash
# Install required tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
export PATH=$PATH:$(go env GOPATH)/bin

# Security scan (takes ~2 seconds) - NEVER CANCEL
make sec
# OR directly: gosec -conf .gosec.json ./...

# Basic formatting checks (takes <1 second)
gofmt -d -s .

# Go static analysis (takes ~2 seconds) - NEVER CANCEL
go vet ./...

# Import formatting
go install golang.org/x/tools/cmd/goimports@latest
goimports -d .
```

**CRITICAL LINTING ISSUE**: The project uses Go 1.25.1, but golangci-lint is not yet compatible with this version. The CI uses golangci-lint-action@v8 which handles this compatibility issue. For local development:
- `make lint` will fail with "Go language version used to build golangci-lint is lower than the targeted Go version"
- Use `gofmt`, `goimports`, `go vet`, and `gosec` instead for local validation
- The CI pipeline will run the full linting suite successfully

## Validation

### Example Application Testing
```bash
# Build example (takes <1 second)
cd examples && go build -o example main.go

# Test example with Docker RabbitMQ (takes ~30 seconds total)
# NEVER CANCEL - Docker pull may take 20+ seconds on first run
docker run -d --name test-rabbitmq -p 5672:5672 rabbitmq:3-management

# Wait for RabbitMQ to start (takes ~10 seconds)
sleep 10

# Test the example application (will run until terminated)
timeout 10s ./example
# Should show successful connection, topology creation, and message consumption

# Cleanup
docker stop test-rabbitmq && docker rm test-rabbitmq
```

### Manual Testing Scenarios
Always test library functionality after making changes by creating test scripts similar to the examples directory:

1. **Connection and Topology**: Verify connection establishment and queue/exchange creation
2. **Publishing**: Test message publishing with different routing keys
3. **Consuming**: Test message consumption with typed handlers
4. **Error Handling**: Test retry mechanisms and dead letter queue behavior
5. **Reconnection**: Test automatic reconnection by stopping/starting RabbitMQ

## Common Tasks

### Repository Structure
```
/home/runner/work/bunmq/bunmq/
├── README.md              # Comprehensive documentation
├── CONTRIBUTING.md        # Contribution guidelines
├── Makefile              # Build targets: test, sec, lint, install
├── go.mod                # Go 1.25.1 module definition
├── *.go                  # Library source files
├── examples/main.go      # Example usage application
└── .github/
    └── workflows/
        └── pipeline.yml  # CI: tests, linting, security, releases
```

### Key Components
- **ConnectionManager**: Handles automatic reconnection with exponential backoff
- **Topology**: Declarative infrastructure (exchanges, queues, bindings)
- **Publisher**: Message publishing with deadline support
- **Dispatcher**: Type-safe message consumption with retry/DLQ
- **QueueDefinition**: Fluent API for queue configuration

### Code Style Guidelines
- Follow standard Go idioms and conventions
- Use the existing logging pattern with logrus
- Maintain backward compatibility for public APIs
- Add comprehensive error handling
- Include OpenTelemetry tracing context where applicable

### Build Timing Expectations
- **go mod download**: ~2 seconds
- **go build ./...**: <1 second  
- **go test ./...**: ~14 seconds - NEVER CANCEL
- **go vet ./...**: ~2 seconds - NEVER CANCEL
- **gosec scan**: ~2 seconds - NEVER CANCEL
- **Docker RabbitMQ setup**: ~30 seconds first time, ~10 seconds subsequent - NEVER CANCEL

### CI Pipeline Validation
Always run these commands before committing:
```bash
# Core validation (takes ~18 seconds total)
go mod download
go build ./...
go test ./...
go vet ./...
make sec
gofmt -d -s .
```

The CI pipeline (`.github/workflows/pipeline.yml`) runs:
1. golangci-lint (with Go version compatibility handling)
2. gosec security scanning with SARIF upload
3. Automated versioning and releases on main branch

### Dependencies and External Services
- **RabbitMQ**: Required for full functionality testing
- **Docker**: Recommended for local RabbitMQ testing
- **No external APIs**: Library is self-contained
- **Go modules**: All dependencies managed via go.mod

### Common File Locations
- **Core logic**: `*.go` files in root directory
- **Examples**: `examples/main.go` 
- **Configuration**: `.gosec.json`, `go.mod`, `Makefile`
- **Documentation**: `README.md`, `CONTRIBUTING.md`
- **CI/CD**: `.github/workflows/pipeline.yml`

## Key Differences from Application Development
- This is a **library**, not an executable application
- No main application to "run" - test via examples and consumer applications
- Focus on API design and backward compatibility
- Validation requires creating test scenarios
- Integration testing requires external RabbitMQ instance
- No deployment - only package publication to Go modules

## Error Patterns to Watch For
- Connection string format: `amqp://user:pass@host:port/vhost`
- Topology must be applied before creating dispatchers/publishers
- Publisher needs ConnectionManager, not raw channel
- Context cancellation handling in long-running consumers
- Proper resource cleanup (defer manager.Close())