# migrate-modules-drop-gin.ps1
# 直接删除所有 module.go 中未使用的 gin import。
# 判定标准：文件中不含 `gin.RouterGroup` 或 `gin.`（限定符）的实际使用。

$root = "d:\work\xin\XinFramework\server\apps"

Get-ChildItem -Path $root -Recurse -Filter "module.go" | ForEach-Object {
    $path = $_.FullName
    $text = Get-Content -Path $path -Raw -Encoding UTF8

    if ($text -notmatch 'import\s+"github\.com/gin-gonic/gin"') { return }

    # 是否还在用 gin.* 标识符（不在 import 行内）
    $withoutImport = [regex]::Replace($text, '^\s*import\s+"github\.com/gin-gonic/gin"\s*\r?\n', '', 'Multiline')
    $stillUses = $withoutImport -match '\bgin\.[A-Z]'

    if (-not $stillUses) {
        Write-Host "drop gin import: $path"
        $new = [regex]::Replace($text, '^\s*import\s+"github\.com/gin-gonic/gin"\s*\r?\n', '', 'Multiline')
        Set-Content -Path $path -Value $new -Encoding UTF8 -NoNewline
    }
}