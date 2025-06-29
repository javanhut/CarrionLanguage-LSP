-- Carrion language plugin entry point
-- This file ensures the plugin loads correctly with lazy.nvim

if vim.g.loaded_carrion then
  return
end
vim.g.loaded_carrion = 1

-- Register carrion filetype if not already done
vim.filetype.add({
  extension = {
    crl = 'carrion',
  },
  filename = {
    ['.carrionrc'] = 'carrion',
  },
  pattern = {
    ['.*%.crl'] = 'carrion',
  },
})

-- Create user commands
vim.api.nvim_create_user_command('CarrionLspInfo', function()
  vim.cmd('LspInfo')
end, { desc = 'Show Carrion LSP information' })

vim.api.nvim_create_user_command('CarrionLspRestart', function()
  vim.cmd('LspRestart carrion')
end, { desc = 'Restart Carrion LSP server' })

vim.api.nvim_create_user_command('CarrionHealth', function()
  require('carrion').check()
end, { desc = 'Check Carrion LSP health' })