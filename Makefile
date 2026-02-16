.PHONY: build test lint fmt vet clean ci website

VERSION ?= dev

# Build the gogrid CLI binary
build:
	go build -ldflags "-X github.com/lonestarx1/gogrid/internal/cli.Version=$(VERSION)" -o bin/gogrid ./cmd/gogrid

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

# Build the website (Next.js static export)
website:
	cd website && yarn build

# Run the same checks as GitHub CI (build + vet + test + lint + website)
ci: build vet test lint website

# Remove build artifacts
clean:
	rm -rf bin/
