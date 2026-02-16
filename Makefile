.PHONY: build test lint fmt vet clean ci

# Build the gogrid CLI binary
build:
	go build -o bin/gogrid ./cmd/gogrid

# Run all tests
test:
	go test -race ./...

# Run golangci-lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run ./...

# Format all Go source files
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Run the same checks as GitHub CI (build + vet + test + lint)
ci: build vet test lint

# Remove build artifacts
clean:
	rm -rf bin/
