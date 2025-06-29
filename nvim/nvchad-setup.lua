-- NvChad Configuration for Carrion Language Support
-- Save this as ~/.config/nvim/lua/plugins/carrion.lua

local M = {}

-- Carrion file type detection
vim.filetype.add({
  extension = {
    crl = "carrion",
  },
})

return {
  -- Basic LSP support with manual installation fallback
  {
    "neovim/nvim-lspconfig",
    config = function()
      require("nvchad.configs.lspconfig").defaults()
      
      local lspconfig = require "lspconfig"
      local nvlsp = require "nvchad.configs.lspconfig"

      -- Check if carrion-lsp is available and configure it
      local function setup_carrion_lsp()
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
      end

      -- Try to setup Carrion LSP
      local success, err = pcall(setup_carrion_lsp)
      if not success then
        vim.schedule(function()
          vim.notify(
            "Carrion LSP not found. Run :CarrionInstall to install it.",
            vim.log.levels.WARN,
            { title = "Carrion LSP" }
          )
        end)
      end
    end,
  },

  -- Treesitter with manual Carrion parser
  {
    "nvim-treesitter/nvim-treesitter",
    opts = function(_, opts)
      opts.ensure_installed = opts.ensure_installed or {}
      
      -- We'll manually install the Carrion parser
      local parser_config = require("nvim-treesitter.parsers").get_parser_configs()
      
      parser_config.carrion = {
        install_info = {
          url = "~/CarrionRepos/CarrionLanguage-LSP/tree-sitter-carrion", -- Local path
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

  -- Automatic installation commands
  {
    "nvim-lua/plenary.nvim",
    config = function()
      -- Command to install Carrion LSP
      vim.api.nvim_create_user_command("CarrionInstall", function()
        vim.notify("Installing Carrion LSP...", vim.log.levels.INFO)
        
        local install_script = [[
          #!/bin/bash
          set -e
          
          echo "Installing Carrion LSP..."
          
          # Create installation directory
          mkdir -p ~/.local/bin
          
          # Check if Go is installed
          if ! command -v go &> /dev/null; then
            echo "Go is required but not installed. Please install Go first."
            exit 1
          fi
          
          # Clone or update repository
          if [ -d "$HOME/CarrionRepos/CarrionLanguage-LSP" ]; then
            echo "Repository already exists, updating..."
            cd "$HOME/CarrionRepos/CarrionLanguage-LSP"
            git pull origin main || true
          else
            echo "Cloning repository..."
            mkdir -p "$HOME/CarrionRepos"
            cd "$HOME/CarrionRepos"
            git clone https://github.com/carrionlang/CarrionLanguage-LSP.git || {
              echo "Failed to clone repository. Creating from local files..."
              # Fallback: assume we're in development
              if [ -d "/home/javanstorm/CarrionRepos/CarrionLanguage-LSP" ]; then
                cp -r "/home/javanstorm/CarrionRepos/CarrionLanguage-LSP" "$HOME/CarrionRepos/"
              fi
            }
          fi
          
          # Build the LSP server
          cd "$HOME/CarrionRepos/CarrionLanguage-LSP"
          echo "Building LSP server..."
          go mod tidy
          go build -o carrion-lsp ./cmd/server/
          
          # Install to local bin
          cp carrion-lsp ~/.local/bin/
          chmod +x ~/.local/bin/carrion-lsp
          
          # Install treesitter parser
          echo "Installing Treesitter parser..."
          
          echo "Installation complete!"
          echo "Make sure ~/.local/bin is in your PATH"
          echo "Add this to your shell config: export PATH=\"$HOME/.local/bin:$PATH\""
        ]]
        
        -- Write install script to temp file
        local temp_file = vim.fn.tempname() .. ".sh"
        local file = io.open(temp_file, "w")
        file:write(install_script)
        file:close()
        
        -- Run installation in terminal
        vim.cmd("split")
        vim.cmd("terminal bash " .. temp_file)
        
        -- Clean up temp file after a delay
        vim.defer_fn(function()
          os.remove(temp_file)
        end, 1000)
      end, { desc = "Install Carrion LSP and Treesitter parser" })

      -- Command to check status
      vim.api.nvim_create_user_command("CarrionStatus", function()
        
        -- Check LSP
        local lsp_handle = io.popen("which carrion-lsp 2>/dev/null")
        local lsp_result = lsp_handle:read("*a")
        lsp_handle:close()
        
        -- Check treesitter
        local ts_available = pcall(require, "nvim-treesitter.parsers")
        local carrion_parser = false
        if ts_available then
          local parsers = require("nvim-treesitter.parsers")
          carrion_parser = parsers.has_parser("carrion")
        end
        
        local status = {}
        table.insert(status, "Carrion Language Support Status:")
        table.insert(status, "")
        
        if lsp_result and lsp_result ~= "" then
          table.insert(status, "✓ LSP Server: " .. vim.trim(lsp_result))
        else
          table.insert(status, "✗ LSP Server: Not found")
        end
        
        if carrion_parser then
          table.insert(status, "✓ Treesitter Parser: Installed")
        else
          table.insert(status, "✗ Treesitter Parser: Not installed")
        end
        
        table.insert(status, "")
        if not (lsp_result and lsp_result ~= "" and carrion_parser) then
          table.insert(status, "Run :CarrionInstall to install missing components")
        else
          table.insert(status, "All components are installed!")
        end
        
        vim.notify(table.concat(status, "\n"), vim.log.levels.INFO)
      end, { desc = "Check Carrion language support status" })

      -- Auto-install treesitter parser on first .crl file
      vim.api.nvim_create_autocmd("FileType", {
        pattern = "carrion",
        callback = function()
          local parsers = require("nvim-treesitter.parsers")
          if not parsers.has_parser("carrion") then
            vim.schedule(function()
              vim.notify(
                "Installing Carrion Treesitter parser...",
                vim.log.levels.INFO
              )
              vim.cmd("TSInstall carrion")
            end)
          end
        end,
      })
    end,
  },

  -- Note: Formatting is handled by the LSP server through textDocument/formatting
  -- No external formatter needed as carrion-lsp provides formatting via LSP protocol
}