-- Carrion Language Server integration for Neovim
-- Optimized for lazy.nvim but works with any plugin manager

local M = {}

-- Plugin information
M.name = 'carrion-lsp'
M.version = '0.1.0'

-- Import modules
local config = require('carrion.config')
local utils = require('carrion.utils')

-- Default configuration
local default_config = {
  cmd = { 'carrion-lsp', '--stdio' },
  filetypes = { 'carrion' },
  root_dir = function(fname)
    local util = require('lspconfig.util')
    return util.root_pattern('.git', '*.crl', 'carrion.toml')(fname) or util.path.dirname(fname)
  end,
  settings = {
    carrion = {
      logLevel = 'info',
      formatting = {
        indentSize = 4,
        insertFinalNewline = true,
      },
    },
  },
  init_options = {},
  capabilities = nil, -- Will be set during setup
}

-- Setup function for the Carrion LSP
function M.setup(opts)
  opts = opts or {}
  
  -- Merge with config defaults
  local final_config = config.setup(opts)
  
  -- Set default capabilities if not provided
  if not final_config.capabilities then
    local has_cmp, cmp_lsp = pcall(require, 'cmp_nvim_lsp')
    if has_cmp then
      final_config.capabilities = cmp_lsp.default_capabilities()
    else
      final_config.capabilities = vim.lsp.protocol.make_client_capabilities()
    end
  end
  
  -- Register the server with lspconfig if available
  local ok, lspconfig = pcall(require, 'lspconfig')
  if ok then
    local configs = require('lspconfig.configs')
    
    -- Register carrion LSP if not already registered
    if not configs.carrion then
      configs.carrion = {
        default_config = vim.tbl_deep_extend('force', default_config, final_config.server or {}),
        docs = {
          description = 'Language Server for the Carrion programming language',
          default_config = {
            root_dir = 'root_pattern(".git", "*.crl")',
          },
        },
      }
    end
    
    -- Setup the server
    lspconfig.carrion.setup(vim.tbl_deep_extend('force', final_config.server or {}, {
      capabilities = final_config.capabilities,
      on_attach = final_config.on_attach,
    }))
  else
    -- Fallback: setup without lspconfig using vim.lsp.start
    vim.api.nvim_create_autocmd('FileType', {
      pattern = 'carrion',
      group = vim.api.nvim_create_augroup('CarrionLsp', { clear = true }),
      callback = function(args)
        local server_config = final_config.server or default_config
        local root_dir = server_config.root_dir and server_config.root_dir(args.file) or vim.fn.getcwd()
        
        vim.lsp.start({
          name = 'carrion-lsp',
          cmd = server_config.cmd,
          root_dir = root_dir,
          capabilities = final_config.capabilities,
          settings = server_config.settings,
          init_options = server_config.init_options,
          on_attach = final_config.on_attach,
        }, {
          bufnr = args.buf,
        })
      end,
    })
  end
  
  -- Ensure filetype detection is set up
  vim.filetype.add({
    extension = {
      crl = 'carrion',
    },
  })
  
  -- Set up additional features
  M._setup_commands()
  M._setup_highlights()
  
  -- Setup integrations
  if final_config.integrations then
    if final_config.integrations.telescope then
      utils.setup_telescope()
    end
  end
  
  if opts.debug then
    print("Carrion LSP configured successfully!")
  end
end

-- Expose utility functions
M.toggle_comment = utils.toggle_comment
M.show_lsp_status = utils.show_lsp_status
M.telescope_carrion_files = utils.telescope_carrion_files

-- Internal function to setup user commands
function M._setup_commands()
  -- Create user commands if they don't exist
  if not vim.api.nvim_get_commands({})['CarrionLspInfo'] then
    vim.api.nvim_create_user_command('CarrionLspInfo', function()
      vim.cmd('LspInfo')
    end, { desc = 'Show Carrion LSP information' })
  end
  
  if not vim.api.nvim_get_commands({})['CarrionLspRestart'] then
    vim.api.nvim_create_user_command('CarrionLspRestart', function()
      vim.cmd('LspRestart carrion')
    end, { desc = 'Restart Carrion LSP server' })
  end
end

-- Internal function to setup additional highlights
function M._setup_highlights()
  -- Define additional highlight groups for better integration
  vim.api.nvim_set_hl(0, 'CarrionKeyword', { link = 'Keyword', default = true })
  vim.api.nvim_set_hl(0, 'CarrionString', { link = 'String', default = true })
  vim.api.nvim_set_hl(0, 'CarrionComment', { link = 'Comment', default = true })
  vim.api.nvim_set_hl(0, 'CarrionNumber', { link = 'Number', default = true })
  vim.api.nvim_set_hl(0, 'CarrionOperator', { link = 'Operator', default = true })
end

-- Health check function
function M.check()
  local health = require('health')
  
  health.start('Carrion LSP')
  
  -- Check if carrion-lsp binary exists
  if vim.fn.executable('carrion-lsp') == 1 then
    health.ok('carrion-lsp binary found in PATH')
    
    -- Try to get version
    local handle = io.popen('carrion-lsp --version 2>&1')
    if handle then
      local version = handle:read('*a')
      handle:close()
      health.info('Version: ' .. vim.trim(version))
    end
  else
    health.error('carrion-lsp binary not found in PATH', {
      'Install carrion-lsp using: make install',
      'Or ensure the binary is in your PATH'
    })
  end
  
  -- Check if lspconfig is available
  if pcall(require, 'lspconfig') then
    health.ok('nvim-lspconfig is available')
  else
    health.warn('nvim-lspconfig not found', {
      'Install nvim-lspconfig for better LSP integration',
      'The LSP will still work without it'
    })
  end
  
  -- Check filetype detection
  vim.cmd('edit test.crl')
  if vim.bo.filetype == 'carrion' then
    health.ok('Filetype detection working')
  else
    health.error('Filetype detection not working')
  end
  vim.cmd('bdelete!')
end

return M