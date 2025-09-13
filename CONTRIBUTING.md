Contributing to bunmq

Thanks for your interest in contributing to bunmq — we appreciate it!

This document explains how you can help improve this project and how we evaluate and accept changes.

Table of contents
- Code of Conduct
- Getting started
- Finding something to work on
- Development environment
- Code style and conventions
- Tests and CI
- Security and static checks
- Pull request process
- Commit message guidelines
- Release & backports
- Contact / maintainers

Code of Conduct
----------------
We expect contributors to follow a friendly and respectful Code of Conduct. Report violations by opening an issue or contacting a maintainer.

Getting started
---------------
1. Fork the repository and create a branch from the default branch (typically `main` or `develop`).
2. Keep changes small and focused. Multiple small PRs are easier and faster to review than a single large one.
3. If the change is non-trivial, open an issue first to discuss the design and approach.

Finding something to work on
----------------------------
- Look for issues labeled `good first issue`, `help wanted`, or `bug`.
- If you want to work on an issue, leave a comment to claim it.
- If none of the labels fit what you want to implement, open an issue describing your idea and link it to your PR.

Development environment
-----------------------
Prerequisites:
- Go (see `go.mod` for the minimum supported version).
- `make` and common developer tools (shell, git).

Common dev commands:

- Build: `make build` or `go build ./...`
- Run unit tests: `make test` or `go test ./...`
- Run linters: `make sec` (gosec) and, if configured, `make lint` (golangci-lint).

If you need a reproducible environment, use a devcontainer (if provided) or Docker.

Code style and conventions
--------------------------
- Follow Go idioms and the official recommendation: https://go.dev/wiki/CodeReviewComments.
- Format code with `gofmt` and keep imports tidy via `goimports` or `gofmt`.
- Keep functions small and tests comprehensive.
- Minimize public API changes; when required, document reasoning in the PR description.

Tests and CI
------------
- Add unit tests for new code and bug fixes.
- Make sure tests pass locally: `go test ./...`.
- If your change affects behavior, add integration or e2e tests when appropriate.
- PRs must pass CI checks (linters, security checks, unit tests) before review and merge.

Security and static checks
--------------------------
- Run `make sec` to execute gosec security checks and address issues.
- Run `make lint` if available to ensure static checks pass.
- If you discover a security issue, follow responsible disclosure: open a private issue or contact a maintainer rather than public disclosure.

Pull request process
--------------------
1. Branch from the default branch and keep it up to date with the base branch.
2. Use a descriptive PR title and include a short summary and motivation in the PR description.
3. Link any relevant issue(s) using `Fixes #NNN` where appropriate.
4. Ensure your branch includes updated `go.mod`/`go.sum` changes if you add/update dependencies.
5. Add tests and update documentation where applicable.
6. Run `make sec` and unit tests locally before pushing.

Review and approval
-------------------
- Maintainers will review your PR. Be responsive to review comments and iterate quickly.
- One or more maintainer approvals may be required before merging.
- Maintainers may squash or rebase commits when merging to keep history clean.

Commit message guidelines
-------------------------
- Use short, imperative subject line (e.g., `publisher: fix nil pointer when...`).
- Provide a longer body when the change is non-trivial explaining the why and high-level what.
- Sign off commits if the project requires DCO (use `git commit --signoff`).

Release & backports
-------------------
- Significant or backward-incompatible changes should be discussed in an issue and coordinated with maintainers.
- Backports to older supported branches must be requested and approved by maintainers.

Contact / maintainers
---------------------
- For questions, open an issue or mention a maintainer in a PR.
- Maintainers are listed in `MAINTAINERS.md` if present.

Thanks for contributing!
