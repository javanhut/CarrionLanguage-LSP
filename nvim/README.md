# Carrion LSP - Neovim Integration

This directory contains ready-to-use Neovim configuration files for the Carrion Language Server.

## Quick Installation

### Option 1: Copy Files Directly

```bash
# Copy all files to your Neovim config directory
cp -r nvim/* ~/.config/nvim/

# Or create symlinks (recommended for development)
ln -s $(pwd)/nvim/* ~/.config/nvim/
```

### Option 2: Plugin Manager Installation

#### With Lazy.nvim

Add to your `~/.config/nvim/lua/plugins/init.lua`:

```lua
{
  dir = "/path/to/CarrionLanguage-LSP/nvim",
  name = "carrion-lsp",
  config = function()
    require("carrion").setup({
      -- Optional configuration
      cmd = { "carrion-lsp", "--stdio" },
      -- Add any custom settings here
    })
  end,
  ft = "carrion",
}
```

#### With Packer.nvim

Add to your `~/.config/nvim/lua/plugins.lua`:

```lua
use {
  "/path/to/CarrionLanguage-LSP/nvim",
  as = "carrion-lsp",
  config = function()
    require("carrion").setup()
  end,
  ft = "carrion",
}
```

#### With vim-plug

Add to your `~/.config/nvim/init.vim`:

```vim
Plug '/path/to/CarrionLanguage-LSP/nvim'
```

Then in Lua configuration:
```lua
require("carrion").setup()
```

## Manual Configuration

If you prefer manual setup, create these files:

### 1. LSP Configuration

`~/.config/nvim/lua/lsp/carrion.lua`:
```lua
local lspconfig = require('lspconfig')
local util = require('lspconfig.util')

-- Register carrion LSP
local configs = require('lspconfig.configs')
if not configs.carrion then
  configs.carrion = {
    default_config = {
      cmd = { 'carrion-lsp', '--stdio' },
      filetypes = { 'carrion' },
      root_dir = util.root_pattern('.git', '*.crl'),
      settings = {},
    },
  }
end

-- Setup with your preferred configuration
lspconfig.carrion.setup({
  on_attach = function(client, bufnr)
    -- Your on_attach function
  end,
  capabilities = require('cmp_nvim_lsp').default_capabilities(),
})
```

### 2. In your `init.lua`:

```lua
-- Load the carrion LSP configuration
require('lsp.carrion')

-- Or use the plugin
require('carrion').setup({
  -- Custom configuration options
})
```

## Features Included

### ✅ Complete LSP Integration
- **Autocompletion**: Keywords, functions, variables, built-ins
- **Diagnostics**: Real-time error detection and warnings
- **Hover Information**: Documentation on symbol hover
- **Go-to-Definition**: Jump to symbol definitions
- **Signature Help**: Function parameter hints
- **Document Formatting**: Auto-format Carrion code

### ✅ Syntax Highlighting
- **Keywords**: `spell`, `grim`, `if`, `for`, etc.
- **Comments**: `#`, `/* */`, ` ``` ` all supported
- **Strings**: Regular, f-strings, and docstrings
- **Numbers**: Integers, floats, hex, binary, octal
- **Operators**: All Carrion operators properly highlighted

### ✅ File Management
- **Filetype Detection**: `.crl` files automatically detected
- **Indentation**: 4-space indentation (configurable)
- **Comment Toggle**: `<leader>/` to toggle comments
- **Folding**: Indent-based folding enabled

### ✅ Key Bindings
- `gd` - Go to definition
- `K` - Show hover information
- `<leader>rn` - Rename symbol
- `<leader>ca` - Code actions
- `gr` - Find references
- `<leader>/` - Toggle comments

## Health Check

Run the health check to verify everything is working:

```vim
:lua require("carrion").check()
```

Or:
```vim
:checkhealth carrion
```

## Troubleshooting

### LSP Not Starting
1. Verify `carrion-lsp` is in PATH: `:!which carrion-lsp`
2. Check LSP status: `:LspInfo`
3. View LSP logs: `:LspLog`

### Syntax Highlighting Issues
1. Check filetype: `:set ft?` (should show `carrion`)
2. Reload syntax: `:syntax off | syntax on`
3. Check syntax file: `:verbose syntax list`

### Completion Not Working
1. Ensure LSP is attached: `:LspInfo`
2. Check completion sources in your completion plugin
3. Try manual completion: `<C-x><C-o>`

## Configuration Options

```lua
require("carrion").setup({
  -- LSP server command
  cmd = { "carrion-lsp", "--stdio" },
  
  -- Additional LSP settings
  settings = {
    carrion = {
      logLevel = "info",
      formatting = {
        indentSize = 4,
        insertFinalNewline = true,
      },
    },
  },
  
  -- Custom on_attach function
  on_attach = function(client, bufnr)
    -- Your custom LSP setup
  end,
})
```

## File Structure

```
nvim/
├── ftdetect/
│   └── carrion.lua          # Filetype detection
├── ftplugin/
│   └── carrion.lua          # Filetype-specific settings
├── syntax/
│   └── carrion.vim          # Syntax highlighting
├── lua/
│   └── carrion/
│       └── init.lua         # Main plugin module
└── README.md                # This file
```

## Requirements

- **Neovim 0.8+** (for modern LSP features)
- **carrion-lsp** binary in PATH
- **nvim-lspconfig** (recommended but not required)

## Optional Enhancements

### With nvim-cmp (Completion)
```lua
require('cmp').setup({
  sources = {
    { name = 'nvim_lsp' },  -- Will include Carrion LSP
    -- other sources...
  }
})
```

### With Telescope (Fuzzy Finding)
```lua
-- Find Carrion files
vim.keymap.set('n', '<leader>fc', function()
  require('telescope.builtin').find_files({
    find_command = { 'fd', '-e', 'crl' }
  })
end)
```

### With Tree-sitter (Enhanced Highlighting)
Note: Tree-sitter support would require a separate grammar. The included syntax file provides comprehensive highlighting for now.

---

**Happy Coding with Carrion in Neovim! 🐦‍⬛✨**