local Pkg = require "mason-core.package"
local _ = require "mason-core.functional"
local platform = require "mason-core.platform"

return Pkg.new {
    name = "carrion-lsp",
    desc = [[Language Server Protocol implementation for the Carrion programming language]],
    homepage = "https://github.com/carrionlang/carrion-lsp",
    languages = { Pkg.Lang.Carrion },
    categories = { Pkg.Cat.LSP },
    ---@async
    ---@param ctx InstallContext
    install = function(ctx)
        local source = ctx:get_github_release_file({
            repo = "carrionlang/carrion-lsp",
            asset_file = _.coalesce(
                _.when(platform.is.mac_x64, "carrion-lsp-darwin-x64.tar.gz"),
                _.when(platform.is.mac_arm64, "carrion-lsp-darwin-arm64.tar.gz"),
                _.when(platform.is.linux_x64, "carrion-lsp-linux-x64.tar.gz"),
                _.when(platform.is.linux_arm64, "carrion-lsp-linux-arm64.tar.gz"),
                _.when(platform.is.win_x64, "carrion-lsp-windows-x64.zip"),
                _.when(platform.is.win_arm64, "carrion-lsp-windows-arm64.zip")
            ),
        })
        
        -- For now, we'll build from source since releases don't exist yet
        ctx:setup_github_repo {
            repo = "carrionlang/carrion-lsp",
            build = function()
                -- Install Go dependencies and build
                ctx:spawn("go", { "mod", "download" })
                ctx:spawn("go", "build", {
                    "-o",
                    platform.is.win and "carrion-lsp.exe" or "carrion-lsp",
                    "./cmd/server",
                })
            end,
        }
        
        ctx:link_bin("carrion-lsp", platform.is.win and "carrion-lsp.exe" or "carrion-lsp")
    end,
}