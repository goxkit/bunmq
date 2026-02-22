# Commit Message Generation Instructions

## Format

Follow the **Conventional Commits** specification. Every commit message MUST use this structure:

```
<type>: <short description>

<body>

<footer>
```

- **Subject line** (`<type>: <short description>`) is REQUIRED.
- **Body** is REQUIRED when the commit touches more than one logical area or more than a few files.
- **Footer** is OPTIONAL (use for breaking changes, issue references).

---

## Subject Line Rules

- Use one of the allowed types listed below — lowercase, no capitalization.
- Scope is OPTIONAL. Use it only when it adds clarity: `feat(publisher): ...`
- Use **imperative mood** ("add", "fix", "remove" — NOT "added", "fixed", "removed").
- Do NOT end with a period.
- Maximum **72 characters**.
- If changes span multiple types, use the **dominant** type (the one that best represents the primary intent of the commit).

---

## Allowed Types

| Type       | When to use                                                        |
|------------|--------------------------------------------------------------------|
| `feat`     | A new feature or capability                                        |
| `fix`      | A bug fix                                                          |
| `docs`     | Documentation only (README, comments, godoc)                       |
| `style`    | Formatting, whitespace, linting — no logic change                  |
| `refactor` | Code restructuring that neither fixes a bug nor adds a feature     |
| `perf`     | Performance improvement                                            |
| `test`     | Adding or updating tests only                                      |
| `build`    | Build system, dependencies, go.mod changes                         |
| `ci`       | CI/CD pipeline configuration (GitHub Actions, Makefile)            |
| `chore`    | Maintenance tasks that don't modify src or test files              |
| `revert`   | Reverting a previous commit                                        |

---

## Body Rules — Grouped Topics

This is the most important rule. When the commit includes changes across **multiple files or logical areas**, the body MUST organize changes into **grouped topics**. Each topic is a short label describing the area or feature, followed by bullet points for the individual changes.

### How to analyze and group

1. **Inspect every staged file diff** — understand what each change does.
2. **Identify logical groups** — cluster related changes by component, feature, or purpose (NOT by filename).
3. **Write a topic header** for each group, followed by concise bullet points.
4. **Use a blank line** between topic groups.
5. **Each bullet should describe _what_ changed and _why_**, not just list filenames.

### Body formatting rules

- Wrap body lines at **80 characters**.
- Use `-` for bullet points (not `*` or `•`).
- Topic headers should be plain text followed by a colon — no markdown formatting.
- Keep bullets concise — one line each when possible.
- Do NOT just list filenames. Describe the semantic change.

---

## Footer Rules

- Use `BREAKING CHANGE: <description>` for breaking API changes.
- Use `Refs: #<number>` to reference related issues.
- Use `Fixes: #<number>` or `Closes: #<number>` to auto-close issues.
- Separate footer from body with a blank line.

---

## Examples

### Single-feature commit (few files)

```
feat: add PublishDeadline method with context timeout

- Accept context with deadline for time-bounded publishing
- Return context.DeadlineExceeded when timeout is reached
- Fall back to default 30s timeout when no deadline is set
```

### Multi-feature commit (many files, grouped topics)

```
feat: add retry mechanism and improve publisher resilience

Publisher:
- Add PublishDeadline method with context timeout support
- Wrap publish errors with structured context for debugging

Connection Manager:
- Implement exponential backoff for reconnection attempts
- Add IsHealthy method with RWMutex protection

Topology:
- Support retry queue declaration with configurable TTL
- Add dead letter exchange binding for failed messages

Tests:
- Add table-driven tests for PublishDeadline edge cases
- Add mock for connection manager reconnection flow
```

### Bug fix with context

```
fix: prevent race condition during channel reconnection

Connection Manager:
- Protect shared channel state with sync.RWMutex
- Use read locks for health checks, write locks for reconnection

Tests:
- Add concurrent access test using goroutines and sync.WaitGroup

Fixes: #123
```

### Chore / dependency update

```
build: update amqp091-go to v1.10.0 and logrus to v1.9.3

- Bump github.com/rabbitmq/amqp091-go from v1.9.0 to v1.10.0
- Bump github.com/sirupsen/logrus from v1.9.2 to v1.9.3
- Run go mod tidy to clean up go.sum
```

### Documentation update

```
docs: add architecture decision records for interface-based design

- Document rationale for interface-based component design
- Add examples of dependency injection patterns used in the library
- Update README with quick-start code snippets
```

### Refactor touching multiple areas

```
refactor: simplify error handling across publisher and dispatcher

Publisher:
- Replace manual error wrapping with fmt.Errorf and %w verb
- Remove redundant nil checks after validated calls

Dispatcher:
- Extract message unmarshaling into dedicated helper function
- Consolidate duplicate error logging into single path

Errors:
- Add ErrInvalidMessageType sentinel error
```

---

## Anti-Patterns — NEVER Do These

- **Generic messages**: "Update files", "Fix stuff", "Various changes", "WIP"
- **File lists without context**: "Update publisher.go, dispatcher.go, mocks.go"
- **Past tense**: "Added feature" — use "add feature"
- **Emojis in the subject line**: No 🚀, ✨, 🐛 in the subject
- **Uppercase type**: "Feat:" or "FIX:" — always lowercase
- **Period at the end**: "fix: resolve connection leak." — no trailing dot
- **Mixing unrelated changes without grouping**: If the body has 10+ bullets with no structure, group them into topics
- **Exceeding 72 chars in the subject**: Keep it concise, move details to the body

---

## Decision Rules for the AI

1. **Always read ALL staged diffs** before writing the commit message.
2. **Count the logical areas touched** — if more than one, use grouped topics in the body.
3. **Pick the dominant type** based on the primary purpose of the commit.
4. **If only one area is changed**, a flat bullet list (no topic headers) is fine.
5. **If tests are the only change**, use `test:` as the type.
6. **If the commit mixes features and tests for those features**, use `feat:` and list tests under a "Tests:" topic group.
7. **Never produce a subject-only commit** when the diff touches 3+ files — always add a body.
