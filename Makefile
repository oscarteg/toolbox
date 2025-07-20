# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=toolbox
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/toolbox/

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Run tests
test:
	$(GOTEST) -v ./...

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/toolbox/

# Install binary to local bin (requires ~/bin in PATH)
install-local: build
	mkdir -p ~/bin
	cp $(BINARY_NAME) ~/bin/
	@echo "Installed to ~/bin/$(BINARY_NAME)"
	@echo "Make sure ~/bin is in your PATH"

# Install binary to /usr/local/bin (system-wide, requires sudo)
install-system: build
	sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

# Install using go install (builds and installs directly)
install-go:
	$(GOCMD) install ./cmd/toolbox/
	@echo "Installed using go install"

# Uninstall from local bin
uninstall-local:
	rm -f ~/bin/$(BINARY_NAME)
	@echo "Removed from ~/bin"

# Uninstall from system bin
uninstall-system:
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Removed from /usr/local/bin"

# Format code
fmt:
	$(GOCMD) fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Development build with race detection
dev:
	$(GOBUILD) -race -o $(BINARY_NAME) -v ./cmd/toolbox/

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  deps          - Install dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  dev           - Build with race detection"
	@echo "  install-local - Install to ~/bin (recommended)"
	@echo "  install-system- Install to /usr/local/bin (requires sudo)"
	@echo "  install-go    - Install using go install"
	@echo "  uninstall-local - Remove from ~/bin"
	@echo "  uninstall-system - Remove from /usr/local/bin"
	@echo "  build-linux   - Cross compile for Linux"

.PHONY: build clean test deps fmt lint dev install-local install-system install-go uninstall-local uninstall-system build-linux help