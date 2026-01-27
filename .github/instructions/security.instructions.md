---
applyTo: "**/*"
---

# Security Instructions

> **Scope**: Applies to all Go files (`**/*.go`), dependency files (`go.mod`, `go.sum`), and container definitions (`Dockerfile`).

## Rule: Dependency Security

**Description**: All dependencies MUST pass `gosec` security checks and use specific versions.

**When it applies**: When adding dependencies or generating code that uses external packages.

**Copilot MUST**:
- Use specific version constraints in `go.mod` (avoid `*` wildcards)
- Prefer well-maintained packages with active security practices
- Check for known CVEs before adding dependencies
- Run `go mod audit` or equivalent security scanning

**Copilot MUST NOT**:
- Add dependencies with known CVEs
- Use `*` version constraints
- Ignore security warnings
- Use abandoned or unmaintained packages

**Example - Correct Dependency Declaration**:

Reference: `@go.mod`

```go
require (
	bitbucket.org/asapcard/go-acquirer-enums-core-lib v0.11.1 // ✅ Specific version
	bitbucket.org/asappay/go-acquirer-messages-lib v1.0.60   // ✅ Specific version
	github.com/go-chi/chi/v5 v5.2.3                          // ✅ Specific version
)
```

**Example - Incorrect Dependency Declaration**:

```go
require (
	github.com/some/package v0.0.0-* // ❌ Wildcard, may include vulnerable versions
	github.com/abandoned/package v1.0.0 // ❌ Abandoned package
)
```

---

## Rule: Never Hardcode Secrets

**Description**: Never hardcode passwords, tokens, API keys, or any sensitive data.

**When it applies**: When generating any code that uses configuration or credentials.

**Copilot MUST**:
- Load secrets from environment variables
- Use configuration management tools (ConfigTool) for sensitive data
- Document required environment variables
- Use remote configuration for customer-specific secrets

**Copilot MUST NOT**:
- Commit secrets to version control
- Hardcode passwords, tokens, or API keys
- Log secret values
- Store secrets in code comments

**Example - Correct Secret Management**:

Reference: `@cmd/container.go`

```go
// ✅ Configuration from environment
configs.LoadEnvironmentProps()

// ✅ Remote configuration for sensitive data
configTool, err := configTool.NewConfigTool()
if err != nil {
    return nil, err
}

// ✅ Customer-specific configuration
customer, err := configs.GetCustomerFromConfigTool(ctx, configTool, authorizedMessage.ClientKey)
```

**Example - Incorrect Secret Management**:

```go
// ❌ WRONG: Hardcoded secrets
const dbPassword = "super_secret_123"
const apiKey = "sk_live_1234567890"

// ❌ WRONG: Secrets in code
func connect() {
    conn, _ := sql.Open("postgres", "user=admin password=secret123")
}
```

---

## Rule: Input Validation

**Description**: Always validate all external input before processing.

**When it applies**: When generating code that receives external input (messages, HTTP requests, configuration).

**Copilot MUST**:
- Validate message structure before processing
- Use message's `Validate()` method when available
- Reject invalid input early (fail fast)
- Validate configuration values at startup

**Copilot MUST NOT**:
- Trust external input without validation
- Process messages without validation
- Skip validation for "performance"

**Example - Correct Input Validation**:

Reference: `@internal/transport/rmq/consumers/authorized.go`

```go
func (c *authorizedConsumer) Handle(ctx context.Context, delivery *amqp091.Delivery) error {
    var authorizedMessage authorizer.AuthorizedMessage
    
    // ✅ Validate unmarshaling
    if err := json.Unmarshal(delivery.Body, &authorizedMessage); err != nil {
        logrus.WithContext(ctx).Error("invalid message, expected authorizedMessage")
        return errors.ErrFailedToUnmarshalJSON
    }
    
    // ✅ Validate message structure
    err := authorizedMessage.Validate()
    if err != nil {
        logrus.WithContext(ctx).WithError(err).Errorf("invalid authorized message")
        return errors.ErrUnformattedMessage
    }
    
    // ✅ Validate customer configuration
    customer, err := configs.GetCustomerFromConfigTool(ctx, c.configTool, authorizedMessage.ClientKey)
    if err != nil {
        logrus.WithContext(ctx).WithError(err).Error("failed to retrieve customer from config tool")
        return errors.ErrConfigTools
    }
    
    return c.authorizerService.Process(ctx, customer, &authorizedMessage)
}
```

---

## Rule: SQL Injection Prevention

**Description**: Always use parameterized queries or prepared statements.

**When it applies**: When generating database access code.

**Copilot MUST**:
- Use parameterized queries with placeholders (`$1`, `$2`, etc.)
- Use database libraries that support parameterization
- Validate input before using in queries
- Use `QueryRowContext` or `ExecContext` with parameters

**Copilot MUST NOT**:
- Concatenate user input into SQL queries
- Use `fmt.Sprintf()` for SQL queries
- Execute raw SQL with user-provided data

**Example - Correct SQL Usage**:

Reference: `@internal/repositories/events_repository.go`

```go
// ✅ Parameterized query
query := `
    SELECT EXISTS(
        SELECT 1
        FROM asap_preagenda
        WHERE nsu = $1
        AND asap_mid = $2
        AND asap_codigo_auth = $3
        AND asap_datavenda = to_timestamp($4)
        AND isactive = 'Y'
    )`

var exists bool
err = db.QueryRowContext(ctx,
    query,
    nsu,                 // ✅ Parameter $1
    mid,                 // ✅ Parameter $2
    authCode,            // ✅ Parameter $3
    authTransactionDate, // ✅ Parameter $4
).Scan(&exists)
```

**Example - Incorrect SQL Usage**:

```go
// ❌ WRONG: String concatenation (SQL injection risk)
query := fmt.Sprintf("SELECT * FROM transactions WHERE nsu = '%s' AND merchant_id = '%s'", nsu, merchantID)
db.Query(query)

// ❌ WRONG: Direct interpolation
query := "SELECT * FROM transactions WHERE nsu = '" + nsu + "'"
```

---

## Rule: Error Information Leakage

**Description**: Never expose internal details, stack traces, or sensitive information in errors returned to external systems.

**When it applies**: When generating error handling code.

**Copilot MUST**:
- Log detailed errors internally with full context
- Return sanitized errors to external systems
- Use error codes for support correlation
- Hide internal paths or configuration details

**Copilot MUST NOT**:
- Expose stack traces externally
- Reveal internal paths or configuration
- Expose existence of resources through errors

**Example - Correct Error Handling**:

```go
// Internal logging with full context
logrus.WithContext(ctx).
    WithError(err).
    WithFields(logrus.Fields{
        "nsu": msg.SystemTraceAuditNumber,
        "merchant_id": msg.MerchantID,
    }).
    Error("failed to process transaction")

// External response - sanitized
return errors.ErrProcessingFailed // ✅ Generic error, no internal details
```

**Example - Incorrect Error Handling**:

```go
// ❌ WRONG: Exposing internal details
return fmt.Errorf("database connection failed: %v, path: /var/lib/postgres/data, config: %+v", err, dbConfig)

// ❌ WRONG: Stack trace in response
return fmt.Errorf("error: %+v", err) // May include stack trace
```

---

## Rule: Secure Configuration

**Description**: Load configuration from environment variables, never hardcode.

**When it applies**: When generating configuration or initialization code.

**Copilot MUST**:
- Load configuration from environment variables
- Use secure configuration management for sensitive data
- Validate configuration values at startup
- Use default values only for non-sensitive settings

**Copilot MUST NOT**:
- Hardcode configuration values
- Store secrets in configuration files committed to version control
- Use insecure defaults for production

**Example - Correct Configuration**:

Reference: `@cmd/container.go`

```go
// ✅ Configuration from environment
configs.LoadEnvironmentProps()
logger.Init(configs.Config.GetString("ENVIRONMENT"), configs.Config.GetString("LOG_LEVEL"))

// ✅ Remote configuration for sensitive data
configTool, err := configTool.NewConfigTool()
```

---

## Rule: Container Security

**Description**: Follow container security best practices.

**When it applies**: When generating Dockerfile or container-related code.

**Copilot MUST**:
- Use minimal base images
- Run containers as non-root user (when possible)
- Scan container images for vulnerabilities
- Implement health checks

**Copilot MUST NOT**:
- Use outdated base images
- Expose unnecessary ports
- Include secrets in container images
- Run as root user unnecessarily

**Example - Correct Dockerfile**:

Reference: `@Dockerfile`

```dockerfile
# ✅ Use specific base image version
FROM golang:1.25.5-alpine AS builder

# ✅ Build stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

# ✅ Minimal runtime image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main", "events"]
```

---

## Security Checklist

When generating code, ensure:

- ✅ No hardcoded secrets
- ✅ All input validated
- ✅ SQL injection prevented (parameterized queries)
- ✅ Path traversal prevented (validate file paths)
- ✅ Errors sanitized (no internal details exposed)
- ✅ Dependencies scanned for vulnerabilities
- ✅ Configuration from environment variables
- ✅ No sensitive data in logs

---

## Rule: Path Traversal Prevention

**Description**: Always validate file paths to prevent path traversal attacks.

**When it applies**: When generating code that handles file paths.

**Copilot MUST**:
- Use `filepath.Clean()` to normalize paths
- Validate paths are within expected directories
- Use `filepath.Join()` for path construction
- Reject paths containing `..` after cleaning

**Copilot MUST NOT**:
- Accept raw user input as file paths
- Use string concatenation for paths
- Trust file paths from external sources

**Example - Correct Path Handling**:

```go
// ✅ Safe path construction
func safePath(baseDir, fileName string) (string, error) {
    cleanName := filepath.Clean(fileName)
    fullPath := filepath.Join(baseDir, cleanName)
    
    // Ensure path is within base directory
    if !strings.HasPrefix(fullPath, filepath.Clean(baseDir)+string(os.PathSeparator)) {
        return "", fmt.Errorf("path traversal attempt detected")
    }
    
    return fullPath, nil
}
```

**Example - Incorrect Path Handling**:

```go
// ❌ WRONG: Direct concatenation
path := baseDir + "/" + userInput

// ❌ WRONG: No validation
file, _ := os.Open(userInput)
```
- ✅ Container security best practices followed
