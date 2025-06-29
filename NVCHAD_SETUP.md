# Carrion Language Support for NvChad

Complete setup guide for Carrion language support in NvChad with LSP, Tree-sitter, and formatting.

## 🚀 Quick Install (Recommended)

For complete beginners, just run this one command:

```bash
curl -fsSL https://raw.githubusercontent.com/carrionlang/carrion-lsp/main/install-nvchad.sh | bash
```

Or download and run locally:

```bash
git clone https://github.com/carrionlang/carrion-lsp.git
cd carrion-lsp
./install-nvchad.sh
```

## 📋 Prerequisites

Before installing, make sure you have:

1. **Neovim** (v0.9+): [Install Guide](https://github.com/neovim/neovim/wiki/Installing-Neovim)
2. **NvChad**: [Install Guide](https://nvchad.com/docs/quickstart/install)
3. **Go** (v1.19+): [Install Guide](https://golang.org/doc/install)
4. **Git**: Usually pre-installed on Linux/macOS

### Installing NvChad (if not installed)

```bash
# Backup existing config
mv ~/.config/nvim ~/.config/nvim.backup

# Install NvChad
git clone https://github.com/NvChad/NvChad ~/.config/nvim --depth 1
```

## 🔧 Manual Installation

If you prefer to install manually or the automatic script doesn't work:

### Step 1: Install Prerequisites

#### Install Go (if not installed)
```bash
# Linux
wget https://go.dev/dl/go1.21.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# macOS
brew install go

# Windows
# Download and install from https://golang.org/dl/
```

### Step 2: Build Carrion LSP

```bash
# Create directory
mkdir -p ~/CarrionRepos
cd ~/CarrionRepos

# Clone repository
git clone https://github.com/carrionlang/carrion-lsp.git CarrionLanguage-LSP
cd CarrionLanguage-LSP

# Build LSP server
go mod tidy
go build -o carrion-lsp ./cmd/server/

# Install to local bin
mkdir -p ~/.local/bin
cp carrion-lsp ~/.local/bin/
chmod +x ~/.local/bin/carrion-lsp

# Add to PATH (add this to your ~/.bashrc or ~/.zshrc)
export PATH="$HOME/.local/bin:$PATH"
```

### Step 3: Install NvChad Plugin

Create the file `~/.config/nvim/lua/plugins/carrion.lua`:

```lua
-- Carrion Language Support for NvChad
-- Save this as ~/.config/nvim/lua/plugins/carrion.lua

local M = {}

-- Carrion file type detection
vim.filetype.add({
  extension = {
    crl = "carrion",
  },
})

return {
  -- Basic LSP support
  {
    "neovim/nvim-lspconfig",
    config = function()
      require("nvchad.configs.lspconfig").defaults()
      
      local lspconfig = require "lspconfig"
      local nvlsp = require "nvchad.configs.lspconfig"

      -- Configure Carrion LSP
      local configs = require "lspconfig.configs"
      
      if not configs.carrion_lsp then
        configs.carrion_lsp = {
          default_config = {
            cmd = { "carrion-lsp" },
            filetypes = { "carrion" },
            root_dir = function(fname)
              return lspconfig.util.find_git_ancestor(fname) 
                or lspconfig.util.path.dirname(fname)
            end,
            settings = {
              carrion = {
                formatting = { enable = true },
                diagnostics = { enable = true },
                completion = { enable = true },
              },
            },
          },
        }
      end

      -- Setup the LSP
      lspconfig.carrion_lsp.setup {
        on_attach = nvlsp.on_attach,
        on_init = nvlsp.on_init,
        capabilities = nvlsp.capabilities,
      }
    end,
  },

  -- Treesitter with Carrion parser
  {
    "nvim-treesitter/nvim-treesitter",
    opts = function(_, opts)
      local parser_config = require("nvim-treesitter.parsers").get_parser_configs()
      
      parser_config.carrion = {
        install_info = {
          url = "~/CarrionRepos/CarrionLanguage-LSP/tree-sitter-carrion",
          files = {"src/parser.c"},
          branch = "main",
          generate_requires_npm = false,
          requires_generate_from_grammar = false,
        },
        filetype = "carrion",
        used_by = { "crl" },
      }
      
      return opts
    end,
  },

  -- Helpful commands
  {
    "nvim-lua/plenary.nvim",
    config = function()
      -- Command to check status
      vim.api.nvim_create_user_command("CarrionStatus", function()
        local notify = require("nvchad.utils").notify
        
        local lsp_handle = io.popen("which carrion-lsp 2>/dev/null")
        local lsp_result = lsp_handle:read("*a")
        lsp_handle:close()
        
        local status = {}
        table.insert(status, "Carrion Language Support Status:")
        table.insert(status, "")
        
        if lsp_result and lsp_result ~= "" then
          table.insert(status, "✓ LSP Server: " .. vim.trim(lsp_result))
        else
          table.insert(status, "✗ LSP Server: Not found")
          table.insert(status, "  Run: ~/.local/bin/carrion-lsp")
        end
        
        notify(table.concat(status, "\n"), vim.log.levels.INFO)
      end, { desc = "Check Carrion language support status" })
    end,
  },

  -- Formatting support
  {
    "stevearc/conform.nvim",
    opts = function(_, opts)
      opts.formatters_by_ft = opts.formatters_by_ft or {}
      opts.formatters_by_ft.carrion = { "carrion_lsp" }
      
      opts.formatters = opts.formatters or {}
      opts.formatters.carrion_lsp = {
        command = "carrion-lsp",
        args = { "--format" },
        stdin = true,
      }
      
      return opts
    end,
  },
}
```

### Step 4: Install Tree-sitter Parser

```bash
# Install Tree-sitter parser
cd ~/CarrionRepos/CarrionLanguage-LSP/tree-sitter-carrion
npm install  # If you have npm
npx tree-sitter generate
```

### Step 5: Test Installation

1. Restart Neovim
2. Run `:Lazy sync` to install plugins
3. Create a test file: `nvim ~/test.crl`
4. Check status: `:CarrionStatus`

## 🎯 Usage

### File Types

Carrion files use the `.crl` extension and are automatically detected.

### Available Commands

- `:CarrionStatus` - Check installation status
- `:TSInstall carrion` - Install Tree-sitter parser
- `:LspInfo` - Show LSP server status
- `:Format` - Format current file

### LSP Features

- **Autocompletion**: Intelligent code completion
- **Diagnostics**: Real-time error checking
- **Formatting**: Automatic code formatting
- **Go to Definition**: Navigate to symbol definitions
- **Hover Information**: Documentation on hover
- **Syntax Highlighting**: Full Tree-sitter highlighting

### Keybindings (NvChad defaults)

- `gd` - Go to definition
- `gr` - Go to references
- `K` - Show hover information
- `<leader>f` - Format file
- `<leader>ca` - Code actions
- `<leader>rn` - Rename symbol

## 📁 Example Carrion Code

Create a file `~/example.crl`:

```carrion
# Example Carrion program
import "math"

grim Calculator:
    ```
    A simple calculator grimoire for basic arithmetic operations.
    Demonstrates Carrion's object-oriented features.
    ```
    
    init():
        ignore
    
    spell add(x, y):
        ```Add two numbers together```
        return x + y
    
    spell multiply(x, y):
        ```Multiply two numbers```
        return x * y
    
    spell divide(x, y):
        ```Divide two numbers with error handling```
        if y == 0:
            raise "Division by zero"
        return x / y

# Usage example
calc = Calculator()
result = calc.add(10, 5)
print(f"10 + 5 = {result}")

# Error handling
attempt:
    bad_result = calc.divide(10, 0)
ensnare error:
    print(f"Error: {error}")
resolve:
    print("Division completed")
```

## 🔧 Troubleshooting

### LSP Server Not Found

```bash
# Check if carrion-lsp is installed
which carrion-lsp

# If not found, check PATH
echo $PATH | grep -o "$HOME/.local/bin"

# Add to PATH if missing
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Tree-sitter Parser Issues

```bash
# Reinstall parser
:TSUninstall carrion
:TSInstall carrion

# Or manually
cd ~/CarrionRepos/CarrionLanguage-LSP/tree-sitter-carrion
npx tree-sitter generate
```

### Plugin Not Loading

```bash
# Check plugin file exists
ls ~/.config/nvim/lua/plugins/carrion.lua

# Restart Neovim and sync plugins
nvim
:Lazy sync
```

### Build Issues

```bash
# Update Go modules
cd ~/CarrionRepos/CarrionLanguage-LSP
go mod tidy
go mod download

# Rebuild
go build -o carrion-lsp ./cmd/server/
```

## 🆕 For Complete Beginners

If you're new to Neovim and NvChad:

1. **Install Neovim**: Follow the [official guide](https://github.com/neovim/neovim/wiki/Installing-Neovim)
2. **Install NvChad**: Run the [installer](https://nvchad.com/docs/quickstart/install)
3. **Run our installer**: `curl -fsSL https://raw.githubusercontent.com/carrionlang/carrion-lsp/main/install-nvchad.sh | bash`
4. **Learn NvChad**: Check the [NvChad documentation](https://nvchad.com/docs/quickstart/learn-nvim)

## 🤝 Support

- **Issues**: [GitHub Issues](https://github.com/carrionlang/carrion-lsp/issues)
- **Discussions**: [GitHub Discussions](https://github.com/carrionlang/carrion-lsp/discussions)
- **Documentation**: [Carrion Language Docs](https://carrionlang.org/docs)

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.