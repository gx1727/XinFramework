# migrate-pool.ps1
# 批量将 apps/*/module.go 里的 `pool := app.DB` 改为 `pool := app.DB.Raw()`
# 适配 appx.Pool 强类型包装。

$root = "d:\work\xin\XinFramework\server\apps"
$files = Get-ChildItem -Path $root -Recurse -Filter "module.go"

foreach ($f in $files) {
    $path = $f.FullName
    $text = Get-Content -Path $path -Raw -Encoding UTF8
    
    # 1. pool := app.DB → pool := app.DB.Raw()
    $new = $text -replace 'pool\s*:=\s*app\.DB\b(?![\.\w])', 'pool := app.DB.Raw()'
    
    if ($new -ne $text) {
        Set-Content -Path $path -Value $new -Encoding UTF8 -NoNewline
        Write-Host "patched: $path"
    } else {
        Write-Host "skipped (no match): $path"
    }
}