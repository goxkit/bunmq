---
applyTo: "**/*"
---

# AI Behavior Instructions

## Rule: Deterministic Code Generation

**Description**: Copilot MUST generate code that follows ALL rules in this instruction system deterministically.

**When it applies**: When generating any code.

**Copilot MUST**:
- Follow all rules from all instruction files
- Reference real files from codebase using `@path/to/file.go`
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
  1. Security rules (highest priority)
  2. Architecture rules (SOLID, clean architecture)
  3. Language-specific rules (Go idioms)
  4. Coding conventions (naming, formatting)
  5. Observability rules (logging, tracing)
  6. Testing rules

**Copilot MUST NOT**:
- Ignore security rules for convenience
- Violate architecture for simplicity
- Skip quality gates

---

## Rule: File References

**Description**: ALWAYS reference real files from codebase when providing examples.

**When it applies**: When showing code examples or patterns.

**Copilot MUST**:
- Use `@path/to/file.go` syntax for file references
- Reference actual code patterns from the codebase
- Show how existing code follows rules
- Use codebase examples, not generic examples

**Copilot MUST NOT**:
- Provide generic examples without codebase context
- Ignore existing codebase patterns
- Suggest patterns not used in codebase

**Example - Correct File References**:

```
✅ Reference: @internal/services/authorized/service.go
✅ Reference: @cmd/container.go
✅ Reference: @internal/repositories/trx_repository.go
```

---

## Rule: Code Modification Behavior

**Description**: When modifying existing code, preserve existing patterns and architecture.

**When it applies**: When editing or refactoring existing code.

**Copilot MUST**:
- Maintain existing architecture patterns
- Follow existing naming conventions
- Preserve existing error handling patterns
- Keep existing logging patterns
- Maintain existing test structure

**Copilot MUST NOT**:
- Refactor without explicit request
- Change architecture patterns
- Modify working code unnecessarily
- Break existing tests
- Change patterns just for "improvement"

---

## Rule: New Code Creation

**Description**: When creating new code, follow all rules and match existing patterns.

**When it applies**: When generating new files or functions.

**Copilot MUST**:
- Follow clean architecture layers
- Use dependency injection
- Implement interfaces correctly
- Add appropriate logging and tracing
- Write tests for new code
- Follow naming conventions
- Reference similar existing code

**Copilot MUST NOT**:
- Create code that violates architecture
- Skip dependency injection
- Ignore logging requirements
- Skip tests
- Create new patterns without justification

---

## Rule: Code Review Before Suggesting

**Description**: Verify code against ALL applicable rules before suggesting.

**When it applies**: Before suggesting any code.

**Copilot MUST CHECK**:
- Does it follow SOLID principles?
- Does it follow clean architecture?
- Does it use proper error handling?
- Does it include logging?
- Does it include tracing?
- Does it follow naming conventions?
- Does it pass quality gates?
- Does it have tests?
- Does it reference existing patterns?

---

## Rule: Complete Code Suggestions

**Description**: Provide complete, working code that follows all rules.

**When it applies**: When suggesting code implementations.

**Copilot MUST**:
- Provide complete implementations
- Include error handling
- Include logging with context
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

---

## Rule: Question Answering

**Description**: When answering questions, reference actual codebase files and patterns.

**When it applies**: When explaining code or architecture.

**Copilot MUST**:
- Reference real files using `@path/to/file.go`
- Explain how codebase implements patterns
- Show actual code examples from codebase
- Reference relevant instruction files
- Provide codebase-specific context

**Copilot MUST NOT**:
- Provide generic answers without codebase context
- Ignore existing codebase patterns
- Suggest patterns not used in codebase
- Give theoretical answers when codebase examples exist

---

## Rule: Error Prevention

**Description**: Actively prevent common mistakes.

**When it applies**: When generating any code.

**Copilot MUST AVOID**:
- Ignoring errors
- Hardcoding configuration
- Missing context in logs
- Skipping input validation
- Missing error handling
- Forgetting to close resources
- Missing tests
- Violating SOLID principles
- Breaking clean architecture
- Skipping security checks

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
- `.github/copilot-instructions.md`: Main entry point, global rules
- `.github/instructions/project-structure.instructions.md`: Directory structure
- `.github/instructions/architecture.instructions.md`: SOLID, clean architecture
- `.github/instructions/golang.instructions.md`: Go-specific rules, tooling
- `.github/instructions/coding-conventions.instructions.md`: Naming, formatting
- `.github/instructions/observability.instructions.md`: Logging, tracing
- `.github/instructions/security.instructions.md`: Security practices
- `.github/instructions/testing.instructions.md`: Testing patterns
- `.github/instructions/ai-behavior.instructions.md`: This file

---

## Rule: Quality Gate Enforcement

**Description**: Ensure all generated code passes quality gates.

**When it applies**: When generating any code.

**Copilot MUST VERIFY**:
- Code compiles
- Passes golangci-lint
- Passes gosec
- Has tests (for business logic)
- Tests pass
- Follows formatting rules

**Copilot MUST NOT**:
- Suggest code that doesn't compile
- Ignore lint warnings
- Skip security checks
- Suggest untested code for business logic
- Suggest code that fails quality gates

---

## Rule: Documentation Generation

**Description**: Generate appropriate documentation comments.

**When it applies**: When creating exported symbols.

**Copilot MUST**:
- Add package comments
- Document exported functions
- Document exported types
- Explain complex logic
- Reference rule files when relevant

**Copilot MUST NOT**:
- Skip documentation for exported symbols
- Add obvious comments
- Document implementation details unnecessarily

---

## Rule: Refactoring Behavior

**Description**: Preserve functionality and follow all rules when refactoring.

**When it applies**: When refactoring code.

**Copilot MUST**:
- Maintain existing functionality
- Follow all rules
- Update tests if needed
- Preserve architecture
- Update documentation

**Copilot MUST NOT**:
- Break existing functionality
- Violate rules during refactoring
- Skip test updates
- Change architecture unnecessarily

---

## AI Behavior Checklist

When generating code, ensure:

- ✅ All rules followed deterministically
- ✅ Real codebase files referenced
- ✅ Existing patterns matched
- ✅ Quality gates passed
- ✅ Complete implementations provided
- ✅ Error handling included
- ✅ Logging with context included
- ✅ Tests included for business logic
- ✅ Documentation added for exported symbols
- ✅ Security rules followed
- ✅ Architecture patterns maintained

---

## Anti-Patterns for AI (FORBIDDEN)

**Copilot MUST NEVER**:
1. Suggest code that violates any rule
2. Ignore existing codebase patterns
3. Create new patterns without justification
4. Skip error handling
5. Skip logging
6. Skip tests for business logic
7. Suggest code that doesn't compile
8. Ignore security rules
9. Break architecture
10. Provide generic examples without codebase context

---

## Positive AI Behaviors (REQUIRED)

**Copilot MUST ALWAYS**:
1. Reference real codebase files
2. Follow all rules deterministically
3. Provide complete, working code
4. Include error handling
5. Include logging with context
6. Include tests for business logic
7. Explain rule application
8. Verify quality gates
9. Use existing patterns
10. Maintain architecture consistency
