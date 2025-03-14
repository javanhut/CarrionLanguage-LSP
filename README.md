# Carrion Language Server for Neovim

This repository contains a Language Server Protocol (LSP) implementation for the Carrion programming language, along with Neovim configuration files to integrate it with NvChad.

## Features

- Syntax highlighting for `.crl` files
- Code completion
- Hover information for keywords and symbols
- Go-to-definition for symbols
- Document formatting
- Error diagnostics

## Installation

### Step 1: Build the Language Server

1. Clone the repository:

```bash
git clone https://github.com/javanhut/carrionlang-lsp.git
```

2. Go into cloned repo:

```bash
cd carrionlang-lsp/
```

3. Build the language server:

```bash
go build -o carrion-lsp ./cmd/server/main.go
```

4. Make the binary available in your PATH:

```bash
sudo mv carrion-lsp /usr/local/bin/
```

# Alternative Installation


1. Install Carrion LSP using Make

```bash
make install
```

2. Uninstall Carrion Lsp using Make
```bash
make uninstall
```

3. Get Help
```bash
make help
```

* Ensure make is installed.
* Note: You may need elevated privileges like sudo to use make option and make must be installed as well.

### Step 2: Configure Neovim with NvChad

#### LSP Configuration

1. Create or update your LSP configuration file at `~/.config/nvim/lua/custom/configs/lspconfig.lua`:

```lua
local on_attach = require("plugins.configs.lspconfig").on_attach
local capabilities = require("plugins.configs.lspconfig").capabilities

local lspconfig = require("lspconfig")
local util = require "lspconfig/util"

-- Your existing LSP configurations...

-- Properly register Carrion LSP as a custom server
local configs = require('lspconfig.configs')
if not configs.carrion then
  configs.carrion = {
    default_config = {
      cmd = {'carrion-lsp', '--stdio'},
      filetypes = {'carrion'},
      root_dir = util.root_pattern(".git"),
      settings = {},
    },
  }
end

-- Setup the Carrion server
lspconfig.carrion.setup {
  on_attach = on_attach,
  capabilities = capabilities,
}
```

#### Filetype Detection

Create a filetype detection file at `~/.config/nvim/ftdetect/carrion.lua`:

```lua
vim.api.nvim_create_autocmd({"BufRead", "BufNewFile"}, {
  pattern = "*.crl",
  callback = function()
    vim.bo.filetype = "carrion"
  end,
})
```

#### Filetype Plugin

Create a filetype plugin at `~/.config/nvim/ftplugin/carrion.lua`:

```lua
-- ftplugin/carrion.lua
-- Set indentation
vim.bo.expandtab = true
vim.bo.shiftwidth = 4
vim.bo.tabstop = 4
vim.bo.softtabstop = 4

-- Enable LSP features for Carrion files
vim.bo.omnifunc = 'v:lua.vim.lsp.omnifunc'
```

#### Syntax Highlighting

Create a syntax file at `~/.config/nvim/syntax/carrion.vim`:

```vim
if exists("b:current_syntax")
  finish
endif

" Keywords
syntax keyword carrionKeyword spell grim init self if else otherwise
syntax keyword carrionKeyword for in while stop skip ignore return import
syntax keyword carrionKeyword match case attempt resolve ensnare raise as
syntax keyword carrionKeyword arcane arcanespell super check and or not

" Boolean values
syntax keyword carrionBoolean True False None

" Comments
syntax match carrionComment "//.*$"

" Strings
syntax region carrionString start=/"/ skip=/\\"/ end=/"/
syntax region carrionString start=/'/ skip=/\\'/ end=/'/

" Numbers
syntax match carrionNumber "\<\d\+\>"
syntax match carrionNumber "\<\d\+\.\d*\>"

" Highlight links
highlight link carrionKeyword Keyword
highlight link carrionBoolean Boolean
highlight link carrionComment Comment
highlight link carrionString String
highlight link carrionNumber Number

let b:current_syntax = "carrion"
```

## Usage

1. Create a new Carrion file with a `.crl` extension:

```bash
touch example.crl
```

2. Open the file in Neovim:

```bash
nvim example.crl
```

3. Start typing Carrion code. You should have:
   - Syntax highlighting
   - Code completion (trigger with Ctrl+X Ctrl+O or automatically)
   - Hover information (trigger with K on a symbol)
   - Go-to-definition (trigger with gd on a symbol)
   - Document formatting (typically on save or with a key binding)

## Example Carrion Code

```
// Example Carrion code
grim Calculator:
    init:
        self.result = 0
        
    spell add(a, b):
        return a + b
        
    spell subtract(a, b):
        return a - b
        
    spell multiply(a, b):
        return a * b
        
    spell divide(a, b):
        if b == 0:
            raise "Division by zero"
        return a / b
```

## Troubleshooting

If you encounter issues:

1. Check if the language server is running:
```bash
ps aux | grep carrion-lsp
```

2. Enable LSP logging in Neovim:
```lua
vim.lsp.set_log_level("debug")
```

3. View LSP logs:
```
:LspInfo
:lua vim.cmd('e'..vim.lsp.get_log_path())
```

4. Start the language server with logging enabled:
```bash
carrion-lsp --stdio --log=/tmp/carrion-lsp.log
```

## License

[MIT License](LICENSE)

## Acknowledgements

- [Carrion Language](https://github.com/javanhut/TheCarrionLanguage) - The language this server supports
- [go.lsp.dev](https://go.lsp.dev/) - LSP implementation for Go
- [NvChad](https://github.com/NvChad/NvChad) - Neovim configuration framework
