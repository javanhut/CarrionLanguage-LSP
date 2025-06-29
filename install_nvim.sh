#!/bin/bash

# Carrion LSP Neovim Installation Script
set -e

NVIM_CONFIG_DIR="${HOME}/.config/nvim"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NVIM_DIR="${SCRIPT_DIR}/nvim"

echo "🐦‍⬛ Installing Carrion LSP for Neovim..."

# Check if Neovim config directory exists
if [ ! -d "$NVIM_CONFIG_DIR" ]; then
    echo "📁 Creating Neovim config directory: $NVIM_CONFIG_DIR"
    mkdir -p "$NVIM_CONFIG_DIR"
fi

# Function to safely copy files
copy_file() {
    local src="$1"
    local dest="$2"
    local desc="$3"
    
    # Create destination directory if it doesn't exist
    mkdir -p "$(dirname "$dest")"
    
    if [ -f "$dest" ]; then
        echo "⚠️  $desc already exists. Creating backup..."
        cp "$dest" "$dest.backup.$(date +%s)"
    fi
    
    echo "📄 Installing $desc..."
    cp "$src" "$dest"
}

# Install filetype detection
copy_file "$NVIM_DIR/ftdetect/carrion.lua" "$NVIM_CONFIG_DIR/ftdetect/carrion.lua" "filetype detection"

# Install filetype plugin
copy_file "$NVIM_DIR/ftplugin/carrion.lua" "$NVIM_CONFIG_DIR/ftplugin/carrion.lua" "filetype plugin"

# Install syntax highlighting
copy_file "$NVIM_DIR/syntax/carrion.vim" "$NVIM_CONFIG_DIR/syntax/carrion.vim" "syntax highlighting"

# Install Lua module
copy_file "$NVIM_DIR/lua/carrion/init.lua" "$NVIM_CONFIG_DIR/lua/carrion/init.lua" "Lua module"

echo ""
echo "✅ Carrion LSP files installed successfully!"
echo ""
echo "📋 Next steps:"
echo "   1. Make sure carrion-lsp is installed: make install"
echo "   2. Add to your Neovim config:"
echo "      require('carrion').setup()"
echo "   3. Open a .crl file to test the integration"
echo "   4. Run health check: :lua require('carrion').check()"
echo ""
echo "📖 For more details, see: nvim/README.md"
echo ""
echo "Happy coding with Carrion! 🐦‍⬛✨"