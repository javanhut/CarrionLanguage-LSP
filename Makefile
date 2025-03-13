# Makefile for Carrion Language Server Protocol

# Configuration
BINARY_NAME = carrion-lsp
INSTALL_PATH = /usr/local/bin
SRC_DIR = ./cmd/server/
GO = go
GOFLAGS = -ldflags="-s -w"
VERSION = 0.1.0

# Detect OS type
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    BINARY_NAME := $(BINARY_NAME).exe
    RM_CMD = del /Q
    MKDIR_CMD = mkdir
else
    DETECTED_OS := $(shell uname -s)
    RM_CMD = rm -f
    MKDIR_CMD = mkdir -p
endif

# Main targets
.PHONY: all build clean install uninstall help

all: build

# Build the LSP server
build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) $(SRC_DIR)
	@echo "Build complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(RM_CMD) $(BINARY_NAME)
	@echo "Clean complete!"

# Install the LSP server
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	$(MKDIR_CMD) $(INSTALL_PATH) 2>/dev/null || true
	cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete!"

# Uninstall the LSP server
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_PATH)..."
	$(RM_CMD) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstallation complete!"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test ./...
	@echo "Tests complete!"

# Show help
help:
	@echo "Carrion LSP Makefile"
	@echo "===================="
	@echo "Available targets:"
	@echo "  make build    - Build the LSP server binary"
	@echo "  make install  - Build and install the LSP server to $(INSTALL_PATH)"
	@echo "  make uninstall- Remove the LSP server from $(INSTALL_PATH)"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make test     - Run tests"
	@echo "  make help     - Show this help message"
	@echo ""
	@echo "Configuration variables (can be overridden):"
	@echo "  BINARY_NAME   - Binary name (default: $(BINARY_NAME))"
	@echo "  INSTALL_PATH  - Installation path (default: $(INSTALL_PATH))"
	@echo "  VERSION       - Version number (default: $(VERSION))"
