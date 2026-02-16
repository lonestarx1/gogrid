.PHONY: build test lint fmt vet clean

# Build the gogrid CLI binary
build:
	go build -o bin/gogrid ./cmd/gogrid

# Run all tests
test:
	go test ./...

# Run golangci-lint
lint:
	golangci-lint run ./...

# Format all Go source files
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Remove build artifacts
clean:
	rm -rf bin/
