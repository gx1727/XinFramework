# migrate-modules-cleanup-v2.ps1
# 简化版：直接删除每个文件中未使用的 slot 绑定行 + 未使用的 gin import。
# 用 PowerShell 解析每个 module.go 的 RegFn 函数体，扫描实际引用。

$root = "d:\work\xin\XinFramework\server\apps"

Get-ChildItem -Path $root -Recurse -Filter "module.go" | ForEach-Object {
    $path = $_.FullName
    $text = Get-Content -Path $path -Raw -Encoding UTF8
    if ($text -notmatch 'slots\.MustGet') { return }

    # 找出 RegFn 函数体范围
    $m = [regex]::Match($text, 'RegFn: func\(ctx plugin\.Reader, slots plugin\.RouterSlots\) \{')
    if (-not $m.Success) { return }
    $start = $m.Index + $m.Length
    # 用括号匹配找函数体结束
    $depth = 1
    $end = $start
    while ($end -lt $text.Length -and $depth -gt 0) {
        $ch = $text[$end]
        if ($ch -eq '{') { $depth++ }
        elseif ($ch -eq '}') { $depth-- }
        $end++
    }
    $body = $text.Substring($start, $end - $start - 1)

    # 哪些 slot 实际在 body 中被引用（按 word boundary）
    # 先剔除 slot 绑定行，避免把 binding 本身误判为"使用"。
    $bodyClean = [regex]::Replace($body, '^\s*(public|tenant|protected)\s*:=.*$', '', 'Multiline')
    $usedPublic    = $bodyClean -match '(?<![A-Za-z0-9_])public(?![A-Za-z0-9_])'
    $usedTenant    = $bodyClean -match '(?<![A-Za-z0-9_])tenant(?![A-Za-z0-9_])'
    $usedProtected = $bodyClean -match '(?<![A-Za-z0-9_])protected(?![A-Za-z0-9_])'

    # 删除未使用的 slot 绑定行（逐行处理避免删除其它有用行）
    $lines = $text -split "`r?`n"
    $out = New-Object System.Collections.Generic.List[string]
    foreach ($line in $lines) {
        $skip = $false
        if ($line -match '^\s*public := slots\.MustGet\(plugin\.SlotPublic\)\.Group\s*$' -and -not $usedPublic) {
            Write-Host "  drop unused slot public in $path"; $skip = $true
        }
        if ($line -match '^\s*tenant := slots\.MustGet\(plugin\.SlotTenant\)\.Group\s*$' -and -not $usedTenant) {
            Write-Host "  drop unused slot tenant in $path"; $skip = $true
        }
        if ($line -match '^\s*protected := slots\.MustGet\(plugin\.SlotProtected\)\.Group\s*$' -and -not $usedProtected) {
            Write-Host "  drop unused slot protected in $path"; $skip = $true
        }
        if (-not $skip) { $out.Add($line) }
    }

    # 检查整个文件还有没有 `gin.`（除 import 行外）
    $joined = ($out -join "`n")
    $ginUsedOutsideImport = ($joined | Select-String -Pattern '\bgin\.' -AllMatches |
        ForEach-Object { $_.Matches } | ForEach-Object { $_.Value } |
        Where-Object { $_ -ne 'gin-gonic/gin' }) -ne $null
    if ($joined -match 'import\s+"github\.com/gin-gonic/gin"' -and -not $ginUsedOutsideImport) {
        Write-Host "  drop unused gin import in $path"
        $out2 = New-Object System.Collections.Generic.List[string]
        foreach ($line in $out) {
            if ($line -notmatch '^\s*import\s+"github\.com/gin-gonic/gin"\s*$') { $out2.Add($line) }
        }
        $out = $out2
    }

    Set-Content -Path $path -Value ($out -join "`r`n") -Encoding UTF8 -NoNewline
    Write-Host "cleaned: $path"
}