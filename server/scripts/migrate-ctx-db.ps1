# migrate-ctx-db.ps1
# 把 module.go 里的 `p := ctx.DB()` 改为 `p := ctx.DB().Raw()`
# 适配 plugin.AppContext.DB() 现在返回 *appx.Pool

$root = "d:\work\xin\XinFramework\server\apps"
$files = Get-ChildItem -Path $root -Recurse -Filter "module.go"

foreach ($f in $files) {
    $path = $f.FullName
    $text = Get-Content -Path $path -Raw -Encoding UTF8
    
    # 1. `p := ctx.DB()` → `p := ctx.DB().Raw()`
    $new = $text -replace '(\bp\s*:=\s*ctx\.DB\(\))(?![\.\w])', '$1.Raw()'
    
    if ($new -ne $text) {
        Set-Content -Path $path -Value $new -Encoding UTF8 -NoNewline
        Write-Host "patched: $path"
    }
}