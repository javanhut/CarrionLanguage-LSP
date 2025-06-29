-- Configuration options for Carrion LSP
local M = {}

-- Default configuration that can be overridden
M.defaults = {
  -- LSP server configuration
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
  
  -- Lazy.nvim specific options
  lazy = {
    ft = { "carrion" },
    cmd = { "CarrionLspInfo", "CarrionLspRestart", "CarrionHealth" },
    keys = {
      { "gd", function() vim.lsp.buf.definition() end, desc = "Go to definition", ft = "carrion" },
      { "K", function() vim.lsp.buf.hover() end, desc = "Hover documentation", ft = "carrion" },
      { "<leader>rn", function() vim.lsp.buf.rename() end, desc = "Rename symbol", ft = "carrion" },
      { "<leader>ca", function() vim.lsp.buf.code_action() end, desc = "Code action", ft = "carrion" },
      { "gr", function() vim.lsp.buf.references() end, desc = "Find references", ft = "carrion" },
      { "<leader>f", function() vim.lsp.buf.format({ async = true }) end, desc = "Format document", ft = "carrion" },
      { "<leader>/", function() require("carrion.utils").toggle_comment() end, desc = "Toggle comment", ft = "carrion" },
    },
  },
  
  -- UI configuration
  ui = {
    border = "rounded",
    winblend = 10,
  },
  
  -- Integration with other plugins
  integrations = {
    telescope = true,
    nvim_cmp = true,
    trouble = true,
  },
}

-- Merge user config with defaults
function M.setup(opts)
  return vim.tbl_deep_extend("force", M.defaults, opts or {})
end

return M