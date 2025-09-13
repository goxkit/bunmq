.PHONY: install test lint sec sec-report

install:
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0
	@go install github.com/goreleaser/goreleaser/v2@latest
	@curl --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

test-coverage-threshold:
	@echo "Checking test coverage threshold (90%)..."
	@go test -coverprofile=coverage.out ./... > /dev/null
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE < 90" | bc -l) -eq 1 ]; then \
		echo "❌ Test coverage is $$COVERAGE%, which is below the 90% threshold"; \
		exit 1; \
	else \
		echo "✅ Test coverage is $$COVERAGE%, which meets the 90% threshold"; \
	fi

lint:
	@command -v golangci-lint >/dev/null || (echo "Install golangci-lint: https://golangci-lint.run/usage/install/"; exit 1)
	golangci-lint run

# Fail CI on >= medium severity by default
sec:
	gosec -conf .gosec.json ./...

# Generate human-friendly reports (HTML + SARIF)
sec-report:
	mkdir -p reports
	gosec -conf .gosec.json -fmt sarif  -out reports/gosec.sarif ./...
	gosec -conf .gosec.json -fmt html   -out reports/gosec.html ./...
	@echo "Reports: reports/gosec.sarif and reports/gosec.html"