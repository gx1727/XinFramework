# migrate-modules.ps1
# 批量替换 19 个业务 module.go 的 RegFn 签名：把四参数 (ctx, public, tenant, protected)
# 改成两参数 (_ plugin.Reader, slots plugin.RouterSlots)。

$root = "d:\work\xin\XinFramework\server\apps"
$pattern = 'RegFn:\s*func\(\s*([a-zA-Z_]+)\s+(plugin\.Reader|_)\s*,\s*([a-zA-Z_]+)\s+\*gin\.RouterGroup\s*,\s*([a-zA-Z_]+)\s+\*gin\.RouterGroup\s*,\s*([a-zA-Z_]+)\s+\*gin\.RouterGroup\s*\)'
$replacement = 'RegFn: func(_ plugin.Reader, slots plugin.RouterSlots)'

Get-ChildItem -Path $root -Recurse -Filter "module.go" | ForEach-Object {
    $path = $_.FullName
    $content = Get-Content -Path $path -Raw -Encoding UTF8
    if ($content -match $pattern) {
        $new = [regex]::Replace($content, $pattern, $replacement)
        Set-Content -Path $path -Value $new -Encoding UTF8 -NoNewline
        Write-Host "patched: $path"
    } else {
        Write-Host "skipped: $path (no match)"
    }
}