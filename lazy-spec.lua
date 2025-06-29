-- Lazy.nvim plugin specification for Carrion LSP
-- Place this in your lazy.nvim plugins configuration

return {
  -- Option 1: If you have the plugin locally
  {
    name = "carrion-lsp",
    dir = "/path/to/CarrionLanguage-LSP/nvim", -- Update this path
    ft = { "carrion" },
    dependencies = {
      "neovim/nvim-lspconfig", -- Recommended but optional
    },
    config = function()
      require("carrion").setup({
        -- Optional configuration
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
          -- Add your custom LSP keybindings here
          local opts = { buffer = bufnr, silent = true }
          vim.keymap.set('n', 'gd', vim.lsp.buf.definition, opts)
          vim.keymap.set('n', 'K', vim.lsp.buf.hover, opts)
          vim.keymap.set('n', '<leader>rn', vim.lsp.buf.rename, opts)
          vim.keymap.set('n', '<leader>ca', vim.lsp.buf.code_action, opts)
          vim.keymap.set('n', 'gr', vim.lsp.buf.references, opts)
          vim.keymap.set('n', '<leader>f', function()
            vim.lsp.buf.format({ async = true })
          end, opts)
        end,
      })
    end,
    cmd = { "CarrionLspInfo", "CarrionLspRestart", "CarrionHealth" },
  },

  -- Option 2: If you want to use it as a Git repository (future)
  -- {
  --   "your-username/carrion-lsp.nvim", -- Replace with actual repo
  --   ft = { "carrion" },
  --   dependencies = {
  --     "neovim/nvim-lspconfig",
  --   },
  --   config = function()
  --     require("carrion").setup()
  --   end,
  -- },

  -- Option 3: Minimal setup
  -- {
  --   name = "carrion-lsp",
  --   dir = "/path/to/CarrionLanguage-LSP/nvim",
  --   ft = "carrion",
  --   config = true, -- Uses default configuration
  -- },
}