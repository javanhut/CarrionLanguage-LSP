-- Carrion Language Support for NvChad
-- Place this file in ~/.config/nvim/lua/plugins/carrion.lua

return {
  -- Mason LSP installer
  {
    "williamboman/mason.nvim",
    opts = {
      ensure_installed = {
        -- We'll manually install carrion-lsp since it's not in the main registry yet
      },
      registries = {
        "github:mason-org/mason-registry",
        -- Add custom registry for carrion-lsp when available
      },
    },
  },

  -- LSP Configuration
  {
    "neovim/nvim-lspconfig",
    dependencies = {
      "williamboman/mason.nvim",
      "williamboman/mason-lspconfig.nvim",
    },
    config = function()
      require("nvchad.configs.lspconfig").defaults()
      
      local lspconfig = require "lspconfig"
      local nvlsp = require "nvchad.configs.lspconfig"

      -- Configure Carrion LSP
      local configs = require "lspconfig.configs"
      
      -- Define carrion_lsp if not already defined
      if not configs.carrion_lsp then
        configs.carrion_lsp = {
          default_config = {
            cmd = { "carrion-lsp" },
            filetypes = { "carrion" },
            root_dir = lspconfig.util.root_pattern(".git", "*.crl"),
            settings = {
              carrion = {
                formatting = {
                  enable = true,
                  indentSize = 4,
                  insertFinalNewline = true,
                },
                diagnostics = {
                  enable = true,
                  level = "error",
                },
                completion = {
                  enable = true,
                  triggerCharacters = { ".", ":" },
                },
              },
            },
            capabilities = nvlsp.capabilities,
          },
        }
      end

      -- Setup Carrion LSP
      lspconfig.carrion_lsp.setup {
        on_attach = nvlsp.on_attach,
        on_init = nvlsp.on_init,
        capabilities = nvlsp.capabilities,
      }
    end,
  },

  -- Treesitter configuration for Carrion
  {
    "nvim-treesitter/nvim-treesitter",
    opts = {
      ensure_installed = {
        "vim", "lua", "vimdoc", "html", "css", "javascript", "typescript", "tsx",
        "c", "markdown", "markdown_inline", "python", "go", "rust", "json",
        -- Carrion will be added manually since it's custom
      },
      auto_install = true,
      highlight = {
        enable = true,
        use_languagetree = true,
      },
      indent = {
        enable = true,
      },
    },
    config = function(_, opts)
      require("nvim-treesitter.configs").setup(opts)
      
      -- Add Carrion treesitter configuration
      local parser_config = require("nvim-treesitter.parsers").get_parser_configs()
      parser_config.carrion = {
        install_info = {
          url = "https://github.com/carrionlang/carrion-lsp", -- Replace with actual repo
          files = {"src/parser.c"},
          branch = "main",
          generate_requires_npm = false,
          requires_generate_from_grammar = false,
        },
        filetype = "carrion",
      }
    end,
  },

  -- File type detection
  {
    "nvim-treesitter/nvim-treesitter",
    dependencies = {
      {
        "nathom/filetype.nvim",
        opts = {
          overrides = {
            extensions = {
              crl = "carrion",
            },
          },
        },
      },
    },
  },

  -- Mason installer specifically for LSP servers
  {
    "williamboman/mason-lspconfig.nvim",
    opts = {
      -- We'll add carrion_lsp when it's available in mason registry
      ensure_installed = {
        "lua_ls",
        "html",
        "cssls",
        "tsserver",
        "clangd",
        "pyright",
        "gopls",
        "rust_analyzer",
      },
    },
  },

  -- Formatting configuration
  {
    "stevearc/conform.nvim",
    event = "BufWritePre",
    opts = {
      formatters_by_ft = {
        carrion = { "carrion_lsp" },
        lua = { "stylua" },
        css = { "prettier" },
        html = { "prettier" },
        javascript = { "prettier" },
        typescript = { "prettier" },
        python = { "black" },
        go = { "gofmt" },
        rust = { "rustfmt" },
      },
      format_on_save = {
        timeout_ms = 500,
        lsp_fallback = true,
      },
      formatters = {
        carrion_lsp = {
          command = "carrion-lsp",
          args = { "--format", "--stdin" },
          stdin = true,
        },
      },
    },
  },

  -- Auto-completion
  {
    "hrsh7th/nvim-cmp",
    opts = function()
      local cmp = require "cmp"
      local conf = require "nvchad.configs.cmp"
      
      -- Add Carrion-specific completion sources
      table.insert(conf.sources, { name = "nvim_lsp" })
      
      return conf
    end,
  },

  -- Additional Carrion language support
  {
    "nvim-lua/plenary.nvim",
    lazy = false,
    config = function()
      -- Custom Carrion commands and utilities
      vim.api.nvim_create_user_command("CarrionInstallLsp", function()
        local install_cmd = "curl -fsSL https://raw.githubusercontent.com/carrionlang/carrion-lsp/main/install.sh | bash"
        vim.fn.system(install_cmd)
        print("Carrion LSP installation started. Check :messages for details.")
      end, { desc = "Install Carrion LSP server" })

      vim.api.nvim_create_user_command("CarrionLspStatus", function()
        local handle = io.popen("which carrion-lsp")
        local result = handle:read("*a")
        handle:close()
        
        if result and result ~= "" then
          print("Carrion LSP found at: " .. vim.trim(result))
        else
          print("Carrion LSP not found. Run :CarrionInstallLsp to install.")
        end
      end, { desc = "Check Carrion LSP installation status" })
    end,
  },
}