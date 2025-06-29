.PHONY: build build-all clean test install help

# Default target
help:
	@echo "Available targets:"
	@echo "  build      - Build for current platform"
	@echo "  build-all  - Build for all platforms"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  install    - Install locally"
	@echo "  release    - Build release binaries (requires VERSION env var)"

# Build for current platform
build:
	go build -ldflags="-s -w" -o oauth-util .

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o oauth-util-linux-amd64 .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o oauth-util-linux-arm64 .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o oauth-util-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o oauth-util-darwin-arm64 .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o oauth-util-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o oauth-util-windows-arm64.exe .

# Build release binaries with version
release:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required. Usage: make release VERSION=v1.2.3"; exit 1; fi
	@echo "Building release version $(VERSION)..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o oauth-util-linux-amd64 .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o oauth-util-linux-arm64 .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o oauth-util-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o oauth-util-darwin-arm64 .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o oauth-util-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o oauth-util-windows-arm64.exe .
	@echo "Creating archives..."
	tar -czf oauth-util-linux-amd64.tar.gz oauth-util-linux-amd64
	tar -czf oauth-util-linux-arm64.tar.gz oauth-util-linux-arm64
	tar -czf oauth-util-darwin-amd64.tar.gz oauth-util-darwin-amd64
	tar -czf oauth-util-darwin-arm64.tar.gz oauth-util-darwin-arm64
	zip oauth-util-windows-amd64.zip oauth-util-windows-amd64.exe
	zip oauth-util-windows-arm64.zip oauth-util-windows-arm64.exe

# Clean build artifacts
clean:
	rm -f oauth-util oauth-util-* oauth-util-*.tar.gz oauth-util-*.zip

# Run tests
test:
	go test ./...

# Install locally
install:
	go install .

# Run the CLI
run:
	go run .
