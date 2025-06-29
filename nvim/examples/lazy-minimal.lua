-- Minimal lazy.nvim configuration for Carrion LSP
-- Add this to your lazy.nvim plugins setup

return {
  {
    name = "carrion-lsp",
    dir = vim.fn.stdpath("data") .. "/lazy/carrion-lsp", -- or your custom path
    ft = "carrion",
    dependencies = {
      "neovim/nvim-lspconfig", -- Recommended
    },
    config = function()
      require("carrion").setup()
    end,
  },
}