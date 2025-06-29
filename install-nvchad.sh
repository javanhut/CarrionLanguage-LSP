#!/bin/bash

# Carrion Language Support Installer for NvChad
# This script automatically sets up Carrion language support in NvChad

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_command() {
    if command -v "$1" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# Check prerequisites
print_status "Checking prerequisites..."

# Check if nvim is installed
if ! check_command nvim; then
    print_error "Neovim is not installed. Please install Neovim first."
    exit 1
fi

# Check if Go is installed
if ! check_command go; then
    print_error "Go is not installed. Please install Go first."
    echo "Visit: https://golang.org/doc/install"
    exit 1
fi

# Check if git is installed
if ! check_command git; then
    print_error "Git is not installed. Please install Git first."
    exit 1
fi

# Check if NvChad is installed
NVIM_CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/nvim"
if [ ! -f "$NVIM_CONFIG_DIR/init.lua" ] || ! grep -q "NvChad" "$NVIM_CONFIG_DIR/init.lua" 2>/dev/null; then
    print_warning "NvChad doesn't appear to be installed."
    echo "Please install NvChad first: https://nvchad.com/docs/quickstart/install"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

print_success "Prerequisites check completed!"

# Create necessary directories
print_status "Creating directories..."
mkdir -p "$HOME/CarrionRepos"
mkdir -p "$HOME/.local/bin"
mkdir -p "$NVIM_CONFIG_DIR/lua/plugins"

# Clone or update the Carrion LSP repository
print_status "Setting up Carrion LSP repository..."
if [ -d "$HOME/CarrionRepos/CarrionLanguage-LSP" ]; then
    print_status "Repository already exists, updating..."
    cd "$HOME/CarrionRepos/CarrionLanguage-LSP"
    git pull origin main || print_warning "Failed to update repository"
else
    print_status "Cloning Carrion LSP repository..."
    cd "$HOME/CarrionRepos"
    
    # Try to clone from GitHub (may not exist yet)
    if ! git clone https://github.com/carrionlang/CarrionLanguage-LSP.git 2>/dev/null; then
        print_warning "GitHub repository not found. Copying from local development..."
        
        # Fallback: copy from the current directory if we're in development
        SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
        if [ -d "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/go.mod" ]; then
            cp -r "$SCRIPT_DIR" "$HOME/CarrionRepos/CarrionLanguage-LSP"
            print_success "Copied from local development directory"
        else
            print_error "Could not find Carrion LSP source code"
            exit 1
        fi
    fi
fi

# Build the LSP server
print_status "Building Carrion LSP server..."
cd "$HOME/CarrionRepos/CarrionLanguage-LSP"
go mod tidy
go build -o carrion-lsp ./cmd/server/

# Install the LSP server
print_status "Installing LSP server..."
cp carrion-lsp "$HOME/.local/bin/"
chmod +x "$HOME/.local/bin/carrion-lsp"

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    print_warning "~/.local/bin is not in your PATH"
    echo "Add this line to your shell configuration file (~/.bashrc, ~/.zshrc, etc.):"
    echo "export PATH=\"\$HOME/.local/bin:\$PATH\""
    
    # Try to add it automatically
    SHELL_CONFIG=""
    if [ -f "$HOME/.zshrc" ]; then
        SHELL_CONFIG="$HOME/.zshrc"
    elif [ -f "$HOME/.bashrc" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
    fi
    
    if [ -n "$SHELL_CONFIG" ]; then
        read -p "Add to $SHELL_CONFIG automatically? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$SHELL_CONFIG"
            print_success "Added to $SHELL_CONFIG"
            export PATH="$HOME/.local/bin:$PATH"
        fi
    fi
fi

# Install NvChad plugin configuration
print_status "Installing NvChad plugin configuration..."
cp "$HOME/CarrionRepos/CarrionLanguage-LSP/nvim/nvchad-setup.lua" "$NVIM_CONFIG_DIR/lua/plugins/carrion.lua"

# Build and install Tree-sitter parser
print_status "Setting up Tree-sitter parser..."
cd "$HOME/CarrionRepos/CarrionLanguage-LSP/tree-sitter-carrion"

# Install npm dependencies if needed
if [ -f "package.json" ] && check_command npm; then
    npm install
    npx tree-sitter generate
else
    print_warning "npm not found, skipping Tree-sitter generation"
fi

# Create a simple test file
print_status "Creating test file..."
cat > "$HOME/test.crl" << 'EOF'
# Test Carrion file
import "math"

grim Calculator:
    ```Simple calculator grimoire```
    
    init():
        ignore
    
    spell add(x, y):
        return x + y
    
    spell multiply(x, y):
        return x * y

# Test the calculator
calc = Calculator()
result = calc.add(5, 3)
print(f"5 + 3 = {result}")
EOF

print_success "Test file created at ~/test.crl"

# Final verification
print_status "Verifying installation..."

# Check LSP server
if check_command carrion-lsp; then
    print_success "Carrion LSP server installed successfully"
else
    print_error "Carrion LSP server installation failed"
fi

# Check plugin file
if [ -f "$NVIM_CONFIG_DIR/lua/plugins/carrion.lua" ]; then
    print_success "NvChad plugin configuration installed"
else
    print_error "NvChad plugin configuration installation failed"
fi

print_success "Installation completed!"
echo
echo "Next steps:"
echo "1. Restart your shell or run: export PATH=\"\$HOME/.local/bin:\$PATH\""
echo "2. Start Neovim and run: :Lazy sync"
echo "3. Open the test file: nvim ~/test.crl"
echo "4. Check status with: :CarrionStatus"
echo "5. If needed, run: :CarrionInstall from within Neovim"
echo
echo "Available commands in Neovim:"
echo "  :CarrionStatus  - Check installation status"
echo "  :CarrionInstall - Reinstall/update components"
echo "  :TSInstall carrion - Install Tree-sitter parser"
echo
print_success "Enjoy coding in Carrion with full LSP support!"