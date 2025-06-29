-- Filetype plugin for Carrion language
-- Sets up indentation, folding, and LSP features

-- Indentation settings (4 spaces, no tabs)
vim.bo.expandtab = true
vim.bo.shiftwidth = 4
vim.bo.tabstop = 4
vim.bo.softtabstop = 4

-- Comment settings
vim.bo.commentstring = "# %s"

-- Folding settings
vim.wo.foldmethod = "indent"
vim.wo.foldlevel = 99

-- Enable LSP omnifunc
vim.bo.omnifunc = "v:lua.vim.lsp.omnifunc"

-- Local mappings for Carrion files
local opts = { buffer = true, silent = true }

-- Quick comment toggle
vim.keymap.set("n", "<leader>/", function()
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
end, opts)

-- LSP-specific keybindings (will work when LSP is attached)
vim.keymap.set("n", "gd", vim.lsp.buf.definition, opts)
vim.keymap.set("n", "K", vim.lsp.buf.hover, opts)
vim.keymap.set("n", "<leader>rn", vim.lsp.buf.rename, opts)
vim.keymap.set("n", "<leader>ca", vim.lsp.buf.code_action, opts)
vim.keymap.set("n", "gr", vim.lsp.buf.references, opts)

