# migrate-strongid.ps1
# 把 auth 模块里残留的 .Uint() / .Valid() / .String() 调用全部替换
# alias 类型不能定义 method，需要手动展开转换

$root = "d:\work\xin\XinFramework\server\apps\boot\auth"
$files = Get-ChildItem -Path $root -Recurse -Filter "*.go" | Where-Object { $_.Name -ne "*_test.go" }

foreach ($f in $files) {
    $path = $f.FullName
    $text = Get-Content -Path $path -Raw -Encoding UTF8
    $original = $text

    # .Uint() 调用替换为 uint(...) 转换
    $text = $text -replace '([a-zA-Z_][a-zA-Z0-9_]*)\.Uint\(\)', 'uint($1)'

    # .Valid() 替换为 != 0
    $text = $text -replace '([a-zA-Z_][a-zA-Z0-9_]*)\.Valid\(\)', '$1 != 0'

    # GetSessionID().String() 替换为 string(GetSessionID())
    $text = $text -replace '([a-zA-Z_][a-zA-Z0-9_]*)\.GetSessionID\(\)\.String\(\)', 'string($1.GetSessionID())'

    if ($text -ne $original) {
        Set-Content -Path $path -Value $text -Encoding UTF8 -NoNewline
        Write-Host "patched: $path"
    }
}