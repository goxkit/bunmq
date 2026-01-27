---
applyTo: "**/*_test.go"
---

# Testing Instructions

## Rule: Test File Naming

**Description**: Test files MUST be named with `*_test.go` suffix and placed in the same package.

**When it applies**: When creating test files.

**Copilot MUST**:
- Name test files with `*_test.go` suffix
- Place test files in the same package as the code being tested
- Use descriptive test function names: `Test<FunctionName>_<Scenario>`

**Copilot MUST NOT**:
- Create test files without `_test.go` suffix
- Put tests in separate packages unless testing exported API
- Use generic test function names

**Example - Correct Test File Naming**:

```
internal/services/authorized/
├── service.go          # Implementation
└── service_test.go     # ✅ Tests in same package
```

---

## Rule: Table-Driven Tests

**Description**: Use table-driven tests for multiple test cases.

**When it applies**: When testing functions with multiple scenarios.

**Copilot MUST**:
- Use table-driven tests for similar test cases
- Structure test tables with clear field names
- Use subtests with `t.Run()` for each test case
- Include setup, execution, and assertion in each test case

**Copilot MUST NOT**:
- Duplicate test code for similar cases
- Create separate test functions for similar scenarios
- Mix different test patterns in one file

**Example - Correct Table-Driven Test**:

```go
func TestAuthorizedService_Process(t *testing.T) {
    tests := []struct {
        name           string
        setupMocks     func(*mocks.MockEventsRepository, *mocks.MockOperationFactory)
        message        *authorizer.AuthorizedMessage
        wantErr        bool
        expectedErr    error
    }{
        {
            name: "successful processing",
            setupMocks: func(repo *mocks.MockEventsRepository, factory *mocks.MockOperationFactory) {
                repo.On("WasProcessed", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
                    Return(false, nil)
                factory.On("GetOperation", mock.Anything).
                    Return(&mocks.MockOperation{}, nil)
            },
            message:     &authorizer.AuthorizedMessage{SystemTraceAuditNumber: "123456"},
            wantErr:     false,
            expectedErr: nil,
        },
        {
            name: "duplicate event skipped",
            setupMocks: func(repo *mocks.MockEventsRepository, factory *mocks.MockOperationFactory) {
                repo.On("WasProcessed", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
                    Return(true, nil) // Already processed
            },
            message:     &authorizer.AuthorizedMessage{SystemTraceAuditNumber: "123456"},
            wantErr:     false, // Should skip, not error
            expectedErr: nil,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockRepo := mocks.NewMockEventsRepository(t)
            mockFactory := mocks.NewMockOperationFactory(t)
            if tt.setupMocks != nil {
                tt.setupMocks(mockRepo, mockFactory)
            }
            
            service := authorized.NewAuthorizedService(mockFactory, mockRepo)
            
            // Execute
            err := service.Process(context.Background(), &configuration.CustomerConfig{}, tt.message)
            
            // Assert
            if (err != nil) != tt.wantErr {
                t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
            }
            if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
                t.Errorf("Process() error = %v, expected %v", err, tt.expectedErr)
            }
        })
    }
}
```

---

## Rule: Mock Usage

**Description**: Use mocks for external dependencies, generate them using `mockgen`.

**When it applies**: When testing code with external dependencies.

**Copilot MUST**:
- Use mocks for repositories, message producers/consumers, configuration tools
- Generate mocks using `mockgen` or similar tools
- Store mocks in `mocks/` directory
- Use interface-based mocks (not concrete types)

**Copilot MUST NOT**:
- Mock concrete types (mock interfaces)
- Create manual mocks when generators exist
- Test implementation details through mocks

**Example - Correct Mock Usage**:

Reference: `@mocks/operation.go` (if exists)

```go
// Generated mock for interface
type MockOperation struct {
    mock.Mock
}

func (m *MockOperation) Process(ctx context.Context, customer *configuration.CustomerConfig, msg *authorizer.AuthorizedMessage) error {
    args := m.Called(ctx, customer, msg)
    return args.Error(0)
}
```

**Test Usage**:

```go
func TestVisaCredit_Process(t *testing.T) {
    // ✅ Mock interface, not concrete type
    mockProducer := mocks.NewMockProducer(t)
    mockProducer.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(nil)
    
    operation := visa.NewVisaCredit(mockProducer)
    
    err := operation.Process(context.Background(), customer, &authorizedMessage)
    
    assert.NoError(t, err)
    mockProducer.AssertExpectations(t)
}
```

---

## Rule: Test Coverage Requirements

**Description**: Aim for high test coverage for business logic, repositories, and error handling.

**When it applies**: When generating test code.

**Copilot MUST**:
- Write tests for business logic in services
- Write tests for repository data access
- Write tests for error handling paths
- Write tests for complex algorithms

**Copilot MUST NOT**:
- Test only happy paths
- Ignore error cases
- Skip testing complex logic

**Example - Coverage Command**:

Reference: `@Makefile`

```makefile
cov-report: ## Generate test coverage report
	go test -coverprofile=coverage.out $$(go list ./...)
```

---

## Rule: Test Organization

**Description**: Organize tests to match source code structure and use helper functions.

**When it applies**: When creating test files.

**Copilot MUST**:
- Organize tests to match source code structure
- Group related tests together
- Use helper functions for common test setup
- Clean up test resources (use `defer` or `t.Cleanup()`)

**Example - Correct Test Organization**:

```go
package authorized

import (
    "context"
    "testing"
    
    "bitbucket.org/asappay/go-clearing-events-ms/mocks"
    // ...
)

// Helper function for common setup
func setupTestService(t *testing.T) (*authorizedService, *mocks.MockEventsRepository, *mocks.MockOperationFactory) {
    mockRepo := mocks.NewMockEventsRepository(t)
    mockFactory := mocks.NewMockOperationFactory(t)
    service := NewAuthorizedService(mockFactory, mockRepo)
    return service, mockRepo, mockFactory
}

func TestAuthorizedService_Process(t *testing.T) {
    // Test cases...
}

func TestAuthorizedService_Process_Duplicate(t *testing.T) {
    // Test duplicate handling...
}
```

---

## Rule: Test Data and Fixtures

**Description**: Use test fixtures for complex data structures, place them in `data/` directory.

**When it applies**: When generating test data.

**Copilot MUST**:
- Use test fixtures for complex data structures
- Place test data in `data/` directory when reusable
- Create test data programmatically for simple cases

**Copilot MUST NOT**:
- Hardcode large test data in test files
- Use production data in tests
- Create test data that's difficult to maintain

**Example - Correct Test Data Usage**:

Reference: `@data/credit_input_auth.json` (if exists)

```go
func loadTestMessage(t *testing.T) *authorizer.AuthorizedMessage {
    data, err := os.ReadFile("data/credit_input_auth.json")
    require.NoError(t, err)
    
    var msg authorizer.AuthorizedMessage
    err = json.Unmarshal(data, &msg)
    require.NoError(t, err)
    
    return &msg
}

func TestVisaCredit_Process(t *testing.T) {
    msg := loadTestMessage(t)
    // Use in test...
}
```

**Example - Programmatic Test Data**:

```go
func createTestAuthorizedMessage() *authorizer.AuthorizedMessage {
    return &authorizer.AuthorizedMessage{
        SystemTraceAuditNumber: "123456",
        MerchantID:              "MERCHANT123",
        TransactionType:        enums.TRX_CREDIT,
        Product:                enums.PDT_VISA,
        // ... other fields
    }
}
```

---

## Rule: Test Assertions

**Description**: Use clear assertions with helpful error messages.

**When it applies**: When writing test assertions.

**Copilot MUST**:
- Use `testify/assert` or `testify/require` for assertions
- Provide context in assertion messages
- Use descriptive error messages

**Copilot MUST NOT**:
- Use generic assertions without messages
- Ignore assertion failures
- Use `if` statements instead of assertions

**Example - Correct Assertions**:

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAuthorizedService_Process(t *testing.T) {
    service, mockRepo, mockFactory := setupTestService(t)
    
    // Setup mocks
    mockRepo.On("WasProcessed", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(false, nil)
    
    err := service.Process(context.Background(), customer, &authorizedMessage)
    
    // ✅ Clear assertions with context
    assert.NoError(t, err, "Process should not return error for valid message")
    mockRepo.AssertExpectations(t)
}
```

---

## Testing Checklist

When generating tests, ensure:

- ✅ Test file named `*_test.go`
- ✅ Table-driven tests for multiple cases
- ✅ Mocks used for external dependencies
- ✅ Both success and error paths tested
- ✅ Test data organized and reusable
- ✅ Clear assertions with helpful messages
- ✅ Tests run in parallel when possible
- ✅ Coverage reports generated
- ✅ Test cleanup performed
