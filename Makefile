# Makefile for Carrion Language Server Protocol

# Configuration
BINARY_NAME = carrion-lsp
INSTALL_PATH = /usr/local/bin
SRC_DIR = ./cmd/server/
GO = go
VERSION = 0.1.0
BUILD_DIR = build

# Get git version if available
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags="-s -w -X main.version=$(GIT_VERSION)"

# Detect OS type
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    BINARY_EXT := .exe
    RM_CMD = del /Q
    MKDIR_CMD = mkdir
else
    DETECTED_OS := $(shell uname -s)
    BINARY_EXT :=
    RM_CMD = rm -rf
    MKDIR_CMD = mkdir -p
endif

# Main targets
.PHONY: all build clean install uninstall help test fmt lint deps run build-all release

all: build

# Build the LSP server
build:
	@echo "Building $(BINARY_NAME) version $(GIT_VERSION)..."
	@$(MKDIR_CMD) $(BUILD_DIR) 2>/dev/null || true
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) $(SRC_DIR)
	@echo "Build complete! Binary at $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)"

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@$(MKDIR_CMD) $(BUILD_DIR)/linux 2>/dev/null || true
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/linux/$(BINARY_NAME) $(SRC_DIR)

build-darwin:
	@echo "Building for macOS..."
	@$(MKDIR_CMD) $(BUILD_DIR)/darwin 2>/dev/null || true
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/darwin/$(BINARY_NAME) $(SRC_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/darwin/$(BINARY_NAME)-arm64 $(SRC_DIR)

build-windows:
	@echo "Building for Windows..."
	@$(MKDIR_CMD) $(BUILD_DIR)/windows 2>/dev/null || true
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe $(SRC_DIR)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(RM_CMD) $(BUILD_DIR)
	@echo "Clean complete!"

# Install the LSP server
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@$(MKDIR_CMD) $(INSTALL_PATH) 2>/dev/null || true
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) $(INSTALL_PATH)/$(BINARY_NAME)
	@sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete! $(BINARY_NAME) is now available in your PATH."

# Uninstall the LSP server
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_PATH)..."
	@sudo $(RM_CMD) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstallation complete!"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...
	@echo "Tests complete!"

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Formatting complete!"

# Run linters (requires golangci-lint)
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Skipping linting."; \
		echo "Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s"; \
	fi

# Update dependencies
deps:
	@echo "Updating dependencies..."
	$(GO) mod tidy
	$(GO) mod download
	@echo "Dependencies updated!"

# Run the LSP server in debug mode
run: build
	@echo "Running $(BINARY_NAME) in debug mode..."
	$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) --stdio --log=/tmp/carrion-lsp.log

# Create release packages
release: clean build-all
	@echo "Creating release packages..."
	@$(MKDIR_CMD) $(BUILD_DIR)/release 2>/dev/null || true
	@cd $(BUILD_DIR)/linux && tar -czf ../release/$(BINARY_NAME)-$(GIT_VERSION)-linux-amd64.tar.gz $(BINARY_NAME)
	@cd $(BUILD_DIR)/darwin && tar -czf ../release/$(BINARY_NAME)-$(GIT_VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)
	@cd $(BUILD_DIR)/darwin && tar -czf ../release/$(BINARY_NAME)-$(GIT_VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-arm64
	@cd $(BUILD_DIR)/windows && zip ../release/$(BINARY_NAME)-$(GIT_VERSION)-windows-amd64.zip $(BINARY_NAME).exe
	@echo "Release packages created in $(BUILD_DIR)/release/"

# Show help
help:
	@echo "Carrion LSP Makefile"
	@echo "===================="
	@echo "Available targets:"
	@echo "  make build      - Build the LSP server binary"
	@echo "  make build-all  - Build for all platforms (Linux, macOS, Windows)"
	@echo "  make install    - Build and install the LSP server to $(INSTALL_PATH)"
	@echo "  make uninstall  - Remove the LSP server from $(INSTALL_PATH)"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Run linters (requires golangci-lint)"
	@echo "  make deps       - Update dependencies"
	@echo "  make run        - Run the LSP server in debug mode"
	@echo "  make release    - Create release packages for all platforms"
	@echo "  make help       - Show this help message"
	@echo ""
	@echo "Configuration variables (can be overridden):"
	@echo "  BINARY_NAME   - Binary name (default: $(BINARY_NAME))"
	@echo "  INSTALL_PATH  - Installation path (default: $(INSTALL_PATH))"
	@echo "  VERSION       - Version number (default: $(VERSION))"
