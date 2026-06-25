# migrate-modules-bindings.ps1
# 在每个 module.go 的 RegFn 函数体开头插入 slots → 局部变量的绑定，
# 同时把第一个参数从 `_` 改回 `ctx`（因为大多数模块需要 ctx 调用 Repository）。
# 之后函数体内既有的 public/tenant/protected 等标识符继续可用。

$root = "d:\work\xin\XinFramework\server\apps"
$pattern = '(RegFn: func\()_ plugin\.Reader, slots plugin\.RouterSlots\) \{\r?\n'
$replacement = @'
$1ctx plugin.Reader, slots plugin.RouterSlots) {
			public := slots.MustGet(plugin.SlotPublic).Group
			tenant := slots.MustGet(plugin.SlotTenant).Group
			protected := slots.MustGet(plugin.SlotProtected).Group
'@

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