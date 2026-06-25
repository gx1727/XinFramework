# migrate-modules-newline.ps1
# 修复上一轮脚本引入时丢失换行的问题：在 protected := 行后插入换行。

$root = "d:\work\xin\XinFramework\server\apps"
$pattern = '(protected := slots\.MustGet\(plugin\.SlotProtected\)\.Group)(?![\r\n])'
$replacement = "`$1`r`n"

Get-ChildItem -Path $root -Recurse -Filter "module.go" | ForEach-Object {
    $path = $_.FullName
    $content = Get-Content -Path $path -Raw -Encoding UTF8
    if ($content -match $pattern) {
        $new = [regex]::Replace($content, $pattern, $replacement)
        Set-Content -Path $path -Value $new -Encoding UTF8 -NoNewline
        Write-Host "fixed: $path"
    }
}