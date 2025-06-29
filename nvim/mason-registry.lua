-- Custom Mason registry for Carrion LSP
-- This file should be placed in ~/.local/share/nvim/site/pack/mason/opt/mason-registry/packages/carrion-lsp/

local Pkg = require "mason-core.package"
local std = require "mason-core.managers.std"
local github = require "mason-core.managers.github"

return Pkg.new {
    name = "carrion-lsp",
    desc = [[Language server for the Carrion programming language with formatting, diagnostics, and autocomplete support.]],
    homepage = "https://github.com/carrionlang/carrion-lsp",
    languages = { Pkg.Lang["Carrion"] },
    categories = { Pkg.Cat.LSP },
    ---@async
    ---@param ctx InstallContext
    install = function(ctx)
        -- Since we don't have releases yet, we'll install by building from source
        github.download_release_file({
            repo = "carrionlang/carrion-lsp",
            out_file = "source.tar.gz",
            asset_file = "source.tar.gz",
        }).with_receipt()
        
        std.extract_file("source.tar.gz", ".")
        
        -- Build the LSP server
        ctx.spawn.go { "build", "-o", "carrion-lsp", "./cmd/server" }
        
        -- Make it executable
        ctx.fs:chmod("+x", "carrion-lsp")
        
        -- Link to bin directory
        ctx:link_bin("carrion-lsp", "carrion-lsp")
    end,
}