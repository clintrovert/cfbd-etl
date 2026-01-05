.PHONY:lint

lint:
	@echo "Installing/updating golangci-lint to latest version..."; \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest
	@echo "Linting Go files..."
	@cd cmd/seeder; \
	LINT_OUTPUT=$$($$(go env GOPATH)/bin/golangci-lint run --config=../../.golangci.yml --skip-files='.*_test\.go$$' --skip-dirs='internal/test|internal/examples' ./... 2>&1); \
	NON_TEST_ERRORS=$$(echo "$$LINT_OUTPUT" | grep "\.go:" | grep -v "_test.go:" || true); \
	if [ -n "$$NON_TEST_ERRORS" ]; then \
		echo "$$NON_TEST_ERRORS"; \
		echo "$$LINT_OUTPUT" | tail -1; \
		exit 1; \
	else \
		echo "Linting checks passed"; \
	fi
