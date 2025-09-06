.PHONY: install lint sec sec-report

install:
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0

lint:
	@command -v golangci-lint >/dev/null || (echo "Install golangci-lint: https://golangci-lint.run/usage/install/"; exit 1)
	golangci-lint run --timeout=5mo

# Fail CI on >= medium severity by default
sec:
	gosec -conf .gosec.json ./...

# Generate human-friendly reports (HTML + SARIF)
sec-report:
	mkdir -p reports
	gosec -conf .gosec.json -fmt sarif  -out reports/gosec.sarif ./...
	gosec -conf .gosec.json -fmt html   -out reports/gosec.html ./...
	@echo "Reports: reports/gosec.sarif and reports/gosec.html"