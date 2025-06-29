-- Full-featured lazy.nvim configuration for Carrion LSP
-- Add this to your lazy.nvim plugins setup

return {
  {
    name = "carrion-lsp",
    dir = "/path/to/CarrionLanguage-LSP/nvim", -- Update this path
    ft = { "carrion" },
    cmd = { "CarrionLspInfo", "CarrionLspRestart", "CarrionHealth" },
    dependencies = {
      "neovim/nvim-lspconfig",
      "hrsh7th/nvim-cmp", -- For completion
      "hrsh7th/cmp-nvim-lsp", -- LSP completion source
    },
    keys = {
      { "<leader>cl", function() require("carrion").show_lsp_status() end, desc = "Carrion LSP Status", ft = "carrion" },
      { "<leader>cf", function() require("carrion").telescope_carrion_files() end, desc = "Find Carrion Files", ft = "carrion" },
      { "<leader>c/", function() require("carrion").toggle_comment() end, desc = "Toggle Comment", ft = "carrion" },
    },
    config = function()
      require("carrion").setup({
        -- Server configuration
        server = {
          cmd = { "carrion-lsp", "--stdio" },
          settings = {
            carrion = {
              logLevel = "info",
              formatting = {
                indentSize = 4,
                insertFinalNewline = true,
              },
              completion = {
                enableSnippets = true,
                enableAutoImport = true,
              },
              diagnostics = {
                enableOnType = true,
                enableOnSave = true,
              },
            },
          },
        },
        
        -- Custom on_attach function
        on_attach = function(client, bufnr)
          -- Enable completion triggered by <c-x><c-o>
          vim.api.nvim_buf_set_option(bufnr, 'omnifunc', 'v:lua.vim.lsp.omnifunc')
          
          -- Mappings
          local opts = { buffer = bufnr, silent = true }
          
          -- LSP mappings
          vim.keymap.set('n', 'gD', vim.lsp.buf.declaration, opts)
          vim.keymap.set('n', 'gd', vim.lsp.buf.definition, opts)
          vim.keymap.set('n', 'K', vim.lsp.buf.hover, opts)
          vim.keymap.set('n', 'gi', vim.lsp.buf.implementation, opts)
          vim.keymap.set('n', '<C-k>', vim.lsp.buf.signature_help, opts)
          vim.keymap.set('n', '<leader>wa', vim.lsp.buf.add_workspace_folder, opts)
          vim.keymap.set('n', '<leader>wr', vim.lsp.buf.remove_workspace_folder, opts)
          vim.keymap.set('n', '<leader>wl', function()
            print(vim.inspect(vim.lsp.buf.list_workspace_folders()))
          end, opts)
          vim.keymap.set('n', '<leader>D', vim.lsp.buf.type_definition, opts)
          vim.keymap.set('n', '<leader>rn', vim.lsp.buf.rename, opts)
          vim.keymap.set('n', '<leader>ca', vim.lsp.buf.code_action, opts)
          vim.keymap.set('n', 'gr', vim.lsp.buf.references, opts)
          vim.keymap.set('n', '<leader>f', function()
            vim.lsp.buf.format { async = true }
          end, opts)
          
          -- Diagnostic mappings
          vim.keymap.set('n', '<leader>e', vim.diagnostic.open_float, opts)
          vim.keymap.set('n', '[d', vim.diagnostic.goto_prev, opts)
          vim.keymap.set('n', ']d', vim.diagnostic.goto_next, opts)
          vim.keymap.set('n', '<leader>q', vim.diagnostic.setloclist, opts)
          
          -- Set up autocommands for the buffer
          if client.server_capabilities.documentHighlightProvider then
            vim.api.nvim_create_augroup("CarrionLspDocumentHighlight", { clear = false })
            vim.api.nvim_clear_autocmds({ buffer = bufnr, group = "CarrionLspDocumentHighlight" })
            vim.api.nvim_create_autocmd({ "CursorHold", "CursorHoldI" }, {
              group = "CarrionLspDocumentHighlight",
              buffer = bufnr,
              callback = vim.lsp.buf.document_highlight,
            })
            vim.api.nvim_create_autocmd("CursorMoved", {
              group = "CarrionLspDocumentHighlight",
              buffer = bufnr,
              callback = vim.lsp.buf.clear_references,
            })
          end
        end,
        
        -- UI configuration
        ui = {
          border = "rounded",
          winblend = 10,
        },
        
        -- Enable integrations
        integrations = {
          telescope = true,
          nvim_cmp = true,
          trouble = true,
        },
        
        -- Debug mode
        debug = false,
      })
    end,
  },
}