-- Utility functions for Carrion LSP
local M = {}

-- Toggle comment for current line or visual selection
function M.toggle_comment()
  local line = vim.api.nvim_get_current_line()
  local cursor_pos = vim.api.nvim_win_get_cursor(0)
  
  if line:match("^%s*#") then
    -- Uncomment: remove the first # and optional space
    local new_line = line:gsub("^(%s*)#%s?", "%1")
    vim.api.nvim_set_current_line(new_line)
  else
    -- Comment: add # at the beginning of non-whitespace
    local indent = line:match("^%s*")
    local content = line:sub(#indent + 1)
    if content ~= "" then
      vim.api.nvim_set_current_line(indent .. "# " .. content)
    end
  end
  
  -- Restore cursor position
  vim.api.nvim_win_set_cursor(0, cursor_pos)
end

-- Check if carrion-lsp binary is available
function M.is_lsp_available()
  return vim.fn.executable("carrion-lsp") == 1
end

-- Get carrion-lsp version
function M.get_lsp_version()
  if not M.is_lsp_available() then
    return nil
  end
  
  local handle = io.popen("carrion-lsp --version 2>&1")
  if handle then
    local version = handle:read("*a")
    handle:close()
    return vim.trim(version)
  end
  return nil
end

-- Create floating window for documentation
function M.create_float_win(content, opts)
  opts = opts or {}
  local width = opts.width or math.floor(vim.o.columns * 0.8)
  local height = opts.height or math.floor(vim.o.lines * 0.8)
  
  local buf = vim.api.nvim_create_buf(false, true)
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, content)
  
  local win_opts = {
    relative = "editor",
    width = width,
    height = height,
    row = math.floor((vim.o.lines - height) / 2),
    col = math.floor((vim.o.columns - width) / 2),
    border = opts.border or "rounded",
    style = "minimal",
  }
  
  local win = vim.api.nvim_open_win(buf, true, win_opts)
  
  -- Set buffer options
  vim.api.nvim_buf_set_option(buf, "buftype", "nofile")
  vim.api.nvim_buf_set_option(buf, "swapfile", false)
  vim.api.nvim_buf_set_option(buf, "modifiable", false)
  
  -- Close on escape
  vim.api.nvim_buf_set_keymap(buf, "n", "<Esc>", "<cmd>close<CR>", { noremap = true, silent = true })
  vim.api.nvim_buf_set_keymap(buf, "n", "q", "<cmd>close<CR>", { noremap = true, silent = true })
  
  return buf, win
end

-- Format diagnostic message for display
function M.format_diagnostic(diagnostic)
  local severity_map = {
    [vim.diagnostic.severity.ERROR] = "ERROR",
    [vim.diagnostic.severity.WARN] = "WARN",
    [vim.diagnostic.severity.INFO] = "INFO",
    [vim.diagnostic.severity.HINT] = "HINT",
  }
  
  local severity = severity_map[diagnostic.severity] or "UNKNOWN"
  return string.format("[%s] %s", severity, diagnostic.message)
end

-- Get project root directory
function M.get_project_root()
  local util = require('lspconfig.util')
  local root = util.root_pattern('.git', '*.crl', 'carrion.toml')(vim.fn.expand('%:p'))
  return root or vim.fn.getcwd()
end

-- Setup telescope integration if available
function M.setup_telescope()
  local ok, telescope = pcall(require, "telescope")
  if not ok then
    return false
  end
  
  -- Register custom pickers
  telescope.register_extension("carrion")
  return true
end

-- Carrion-specific telescope pickers
function M.telescope_carrion_files()
  local ok, builtin = pcall(require, "telescope.builtin")
  if not ok then
    vim.notify("Telescope not available", vim.log.levels.ERROR)
    return
  end
  
  builtin.find_files({
    prompt_title = "Carrion Files",
    find_command = { "fd", "-e", "crl" },
    previewer = true,
  })
end

-- Show LSP status in a floating window
function M.show_lsp_status()
  local clients = vim.lsp.get_active_clients({ name = "carrion" })
  if #clients == 0 then
    vim.notify("No Carrion LSP client active", vim.log.levels.WARN)
    return
  end
  
  local content = {}
  for _, client in ipairs(clients) do
    table.insert(content, "Client: " .. client.name)
    table.insert(content, "Status: " .. (client.is_stopped() and "Stopped" or "Running"))
    table.insert(content, "Root Dir: " .. (client.config.root_dir or "Unknown"))
    table.insert(content, "Capabilities:")
    
    if client.server_capabilities.completionProvider then
      table.insert(content, "  ✓ Completion")
    end
    if client.server_capabilities.hoverProvider then
      table.insert(content, "  ✓ Hover")
    end
    if client.server_capabilities.definitionProvider then
      table.insert(content, "  ✓ Go to Definition")
    end
    if client.server_capabilities.documentFormattingProvider then
      table.insert(content, "  ✓ Formatting")
    end
    if client.server_capabilities.signatureHelpProvider then
      table.insert(content, "  ✓ Signature Help")
    end
  end
  
  M.create_float_win(content, { width = 50, height = #content + 2 })
end

return M