# Carrion Language Server

A comprehensive Language Server Protocol (LSP) implementation for the Carrion programming language, providing intelligent code editing features for various editors including VS Code, Neovim, Emacs, and more.

## Features

### Core Functionality
- 🔍 **Syntax Analysis**: Real-time parsing and error detection
- 📝 **Code Completion**: Context-aware suggestions for:
  - Keywords (`spell`, `grim`, `if`, `for`, etc.)
  - Built-in functions (`print`, `len`, `type`, etc.)
  - User-defined functions and grimoires
  - Variables and methods
- 🎯 **Go to Definition**: Jump to symbol declarations
- 💡 **Hover Information**: Detailed documentation on hover
- 🎨 **Code Formatting**: Automatic indentation and style fixes
- ⚠️ **Diagnostics**: Real-time error and warning messages

### Advanced Features
- Symbol table management for cross-file references
- Semantic analysis for type checking
- Support for Carrion's unique syntax (grimoires, spells, etc.)
- Multi-platform support (Linux, macOS, Windows)

## 🚀 Quick Start

### Prerequisites
- **Git** (for cloning the repository)
- **Go 1.23+** (for building from source)
- **TheCarrionLanguage** v0.1.6 or later

### 1. Install LSP Server

**Recommended (one command):**
```bash
git clone https://github.com/javanhut/CarrionLanguage-LSP.git && cd CarrionLanguage-LSP && make install
```

**Alternative methods:**
<details>
<summary>Click to see other installation options</summary>

**Manual build and install:**
```bash
git clone https://github.com/javanhut/CarrionLanguage-LSP.git
cd CarrionLanguage-LSP
make build
sudo cp build/carrion-lsp /usr/local/bin/
```

**Development build:**
```bash
git clone https://github.com/javanhut/CarrionLanguage-LSP.git
cd CarrionLanguage-LSP
go mod tidy
go build -o carrion-lsp ./cmd/server/
```

</details>

### 2. Verify Installation
```bash
carrion-lsp --version
```

### 3. Setup Your Editor
Choose your editor below for specific setup instructions.

### Available Commands
- `make build` - Build the language server
- `make install` - Build and install to system PATH  
- `make uninstall` - Remove from system
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make help` - Show all available commands

## Editor Integration

### 🎯 Neovim with NvChad (Recommended for Beginners)

#### 🚀 One-Command Install

**For complete beginners with NvChad:**
```bash
curl -fsSL https://raw.githubusercontent.com/carrionlang/carrion-lsp/main/install-nvchad.sh | bash
```

**Manual install:**
```bash
git clone https://github.com/javanhut/CarrionLanguage-LSP.git
cd CarrionLanguage-LSP
./install-nvchad.sh
```

**📖 Complete NvChad Setup Guide**: [NVCHAD_SETUP.md](NVCHAD_SETUP.md)

#### 🚀 Advanced/Manual Neovim Setup

**1. Install LSP:**
```bash
git clone https://github.com/javanhut/CarrionLanguage-LSP.git
cd CarrionLanguage-LSP
make install
```

**2. Add to your Neovim config:**

**For lazy.nvim users (most common):**
```lua
-- In your lazy.nvim plugin list:
{
  name = "carrion-lsp", 
  dir = "/path/to/CarrionLanguage-LSP/nvim",  -- Update this path
  ft = "carrion",
  config = function()
    require("carrion").setup()
  end,
}
```

**For other plugin managers or manual setup:**
```bash
# Copy plugin files to your Neovim config
cp -r /path/to/CarrionLanguage-LSP/nvim/* ~/.config/nvim/

# Add to your init.lua:
require("carrion").setup()
```

**3. Verify installation:**
```
:CarrionHealth
```

#### ✅ What You Get Instantly

- **Complete LSP Integration**: Autocompletion, diagnostics, hover, go-to-definition
- **Syntax Highlighting**: All Carrion syntax with magical keywords
- **Smart Indentation**: 4-space indentation with proper folding  
- **Key Bindings**: `gd`, `K`, `<leader>rn`, `<leader>ca`, `<leader>/` (comment toggle)
- **Commands**: `:CarrionLspInfo`, `:CarrionLspRestart`, `:CarrionHealth`

### 💻 VS Code

#### 🚀 Quick Setup

**1. Install LSP server:**
```bash
git clone https://github.com/javanhut/CarrionLanguage-LSP.git
cd CarrionLanguage-LSP
make install
```

**2. Install VS Code extension:**
```bash
# Copy extension to VS Code
cp -r editors/vscode ~/.vscode/extensions/carrion-language-0.1.0

# Restart VS Code
```

**3. Optional configuration in VS Code settings.json:**
```json
{
  "carrion.server.path": "carrion-lsp",
  "carrion.server.logLevel": "info"
}
```

#### ✅ What You Get

- **IntelliSense**: Smart autocompletion for all Carrion syntax
- **Syntax Highlighting**: Beautiful color-coded Carrion code
- **Error Squiggles**: Real-time error detection and warnings
- **Hover Documentation**: Detailed info on functions and variables
- **Go to Definition**: Jump to symbol definitions with F12
- **Auto-formatting**: Format code with Shift+Alt+F

---

### 🔧 Advanced/Alternative Setups

<details>
<summary><strong>Manual Neovim LSP Configuration (Click to expand)</strong></summary>

If you prefer manual setup without the plugin, create this file at `~/.config/nvim/lua/custom/configs/lspconfig.lua`:

```lua
local on_attach = require("plugins.configs.lspconfig").on_attach
local capabilities = require("plugins.configs.lspconfig").capabilities

local lspconfig = require("lspconfig")
local util = require "lspconfig/util"

-- Register Carrion LSP as a custom server
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

**Add filetype detection** (`~/.config/nvim/ftdetect/carrion.lua`):
```lua
vim.api.nvim_create_autocmd({"BufRead", "BufNewFile"}, {
  pattern = "*.crl",
  callback = function()
    vim.bo.filetype = "carrion"
  end,
})
```

</details>

<details>
<summary><strong>Emacs with LSP-mode (Click to expand)</strong></summary>

**1. Install LSP server:**
```bash
git clone https://github.com/javanhut/CarrionLanguage-LSP.git
cd CarrionLanguage-LSP
make install
```

**2. Add to your `~/.emacs.d/init.el`:**
```elisp
(use-package lsp-mode
  :ensure t
  :hook (carrion-mode . lsp)
  :commands lsp)

;; Define carrion-mode
(define-derived-mode carrion-mode fundamental-mode "Carrion"
  "Major mode for Carrion language files."
  (setq comment-start "# ")
  (setq comment-end ""))

;; Associate .crl files with carrion-mode
(add-to-list 'auto-mode-alist '("\\.crl\\'" . carrion-mode))

;; Register Carrion LSP server
(with-eval-after-load 'lsp-mode
  (add-to-list 'lsp-language-id-configuration '(carrion-mode . "carrion"))
  (lsp-register-client
   (make-lsp-client :new-connection (lsp-stdio-connection "carrion-lsp")
                    :major-modes '(carrion-mode)
                    :server-id 'carrion-lsp)))
```

</details>

<details>
<summary><strong>Vim with CoC (Click to expand)</strong></summary>

**1. Install LSP server:**
```bash
git clone https://github.com/javanhut/CarrionLanguage-LSP.git
cd CarrionLanguage-LSP
make install
```

**2. Configure CoC settings** (`~/.config/nvim/coc-settings.json` or `:CocConfig`):
```json
{
  "languageserver": {
    "carrion": {
      "command": "carrion-lsp",
      "args": ["--stdio"],
      "filetypes": ["carrion"],
      "rootPatterns": [".git/", "."],
      "settings": {}
    }
  }
}
```

**3. Add filetype detection** (`~/.config/nvim/ftdetect/carrion.vim`):
```vim
au BufRead,BufNewFile *.crl set filetype=carrion
```

</details>

## Usage and Features

### Basic Usage

1. **Create a Carrion file**:
   ```bash
   touch my_project.crl
   ```

2. **Open in your editor** (VS Code, Neovim, Emacs, etc.)

3. **Start coding with full LSP support**:
   - 🎯 **Auto-completion**: Type `.` after objects or start typing keywords
   - 💡 **Hover documentation**: Hover over symbols for detailed information
   - 🔍 **Go-to-definition**: Jump to symbol definitions instantly
   - ⚡ **Real-time diagnostics**: See syntax errors and warnings as you type
   - 🎨 **Auto-formatting**: Format your code automatically
   - 📝 **Signature help**: See function parameters while typing

### Example Carrion Code

```carrion
# Advanced Carrion example showcasing LSP features
import "math"

grim Calculator:
    ```
    A magical calculator grimoire with enhanced arithmetic capabilities.
    Supports basic operations and advanced mathematical functions.
    ```
    
    init(precision: int = 2):
        self.precision = precision
        self.history = []
    
    spell add(a: float, b: float) -> float:
        ```Add two numbers with precision tracking```
        result = a + b
        self.history.append(f"add({a}, {b}) = {result}")
        return result
    
    spell power(base: float, exponent: float) -> float:
        ```Calculate base raised to the power of exponent```
        if base == 0 and exponent < 0:
            raise "Cannot divide by zero"
        return base ** exponent

# Function with error handling
spell safe_divide(x: float, y: float) -> float:
    attempt:
        if y == 0:
            raise "Division by zero error"
        return x / y
    ensnare error:
        print(f"Error occurred: {error}")
        return 0.0
    resolve:
        print("Division operation completed")

# Usage with full LSP support
calc = Calculator(3)
result = calc.add(10.5, 20.3)  # LSP shows signature help
print(calc.power(2, 8))        # Hover shows documentation
```

### LSP Features Demonstration

- **Completion**: As you type `calc.`, the LSP will suggest `add`, `power`, `precision`, `history`
- **Hover**: Hovering over `Calculator` shows the docstring and type information
- **Diagnostics**: Syntax errors are highlighted in real-time
- **Signature Help**: When typing `calc.add(`, you'll see parameter information
- **Go-to-Definition**: Clicking on `Calculator` jumps to its definition

## Architecture

### LSP Server Components

The Carrion Language Server is built with a modular architecture:

```
carrion-lsp/
├── cmd/server/           # Entry point and CLI handling
├── internal/
│   ├── handler/          # LSP protocol handlers
│   ├── analyzer/         # Language analysis engine
│   ├── symbols/          # Symbol table management
│   ├── formatter/        # Code formatting
│   ├── protocol/         # LSP protocol abstractions
│   ├── langserver/       # Server lifecycle management
│   └── util/             # Utilities and logging
└── editors/              # Editor-specific configurations
```

### Key Features

1. **Real-time Analysis**: Uses TheCarrionLanguage's lexer and parser
2. **Symbol Management**: Comprehensive symbol table with scope tracking
3. **Multi-transport**: Supports both stdio and TCP communication
4. **Extensible**: Modular design for easy feature addition

## Troubleshooting

### Common Issues

1. **Server not starting**:
   ```bash
   # Check if carrion-lsp is in PATH
   which carrion-lsp
   
   # Verify installation
   carrion-lsp --version
   ```

2. **No completions appearing**:
   ```bash
   # Check LSP client logs (editor-specific)
   # For Neovim:
   :LspLog
   
   # For VS Code: View > Output > Carrion Language Server
   ```

3. **Diagnostics not showing**:
   ```bash
   # Test server directly
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | carrion-lsp --stdio
   ```

4. **Debug mode**:
   ```bash
   # Run with detailed logging
   carrion-lsp --stdio --log=/tmp/carrion-lsp-debug.log
   
   # Monitor the log file
   tail -f /tmp/carrion-lsp-debug.log
   ```

### Performance Tips

- **Large files**: The server handles files up to 10MB efficiently
- **Memory usage**: Typical memory usage is 20-50MB per workspace
- **Startup time**: Initial parsing may take 1-2 seconds for large projects

## Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes**
4. **Run tests**: `make test`
5. **Format code**: `make fmt`
6. **Submit a pull request**

### Development Setup

```bash
git clone https://github.com/javanhut/carrionlang-lsp.git
cd carrionlang-lsp
go mod tidy
make build
make test
```

## Roadmap

### Planned Features
- [ ] **Code actions**: Quick fixes and refactoring
- [ ] **Rename symbol**: Workspace-wide symbol renaming
- [ ] **Find references**: Show all usages of a symbol
- [ ] **Document symbols**: Outline view support
- [ ] **Workspace symbols**: Global symbol search
- [ ] **Incremental parsing**: Faster updates for large files
- [ ] **Multi-root workspaces**: Support for complex project structures

### Editor Support
- [x] **VS Code** (extension available)
- [x] **Neovim** (with built-in LSP)
- [x] **Emacs** (with lsp-mode)
- [x] **Vim** (with CoC)
- [ ] **Sublime Text** (LSP plugin)
- [ ] **IntelliJ IDEA** (plugin planned)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgements

- **[TheCarrionLanguage](https://github.com/javanhut/TheCarrionLanguage)** - The magical language this server supports
- **[go.lsp.dev](https://go.lsp.dev/)** - Excellent LSP implementation for Go
- **[Language Server Protocol](https://microsoft.github.io/language-server-protocol/)** - The protocol specification
- **Go Community** - For the robust tooling and ecosystem

---

**Happy Coding with Carrion! 🐦‍⬛✨**
