.PHONY: lint

lint:
	@command -v golangci-lint >/dev/null || (echo "Install golangci-lint: https://golangci-lint.run/usage/install/"; exit 1)
	golangci-lint run --timeout=5m