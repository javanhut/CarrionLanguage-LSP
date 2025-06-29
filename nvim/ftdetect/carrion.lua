-- Filetype detection for Carrion language files
vim.api.nvim_create_autocmd({ "BufRead", "BufNewFile" }, {
	pattern = "*.crl",
	callback = function()
		vim.wo.filetype = "carrion"
	end,
})

