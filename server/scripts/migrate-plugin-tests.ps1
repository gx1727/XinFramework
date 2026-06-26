# migrate-plugin-tests.ps1
# 把 plugin 测试文件里的 `NewAppContext(&pgxpool.Pool{}, ...)` 改为 `NewAppContext(appx.MustNewPool(&pgxpool.Pool{}), ...)`
# 把 `ctx.DB() != pool` 改为 `ctx.DB().Raw() != pool`

$root = "d:\work\xin\XinFramework\server\framework\pkg\plugin"
$files = Get-ChildItem -Path $root -Filter "*_test.go"

foreach ($f in $files) {
    $path = $f.FullName
    $text = Get-Content -Path $path -Raw -Encoding UTF8
    $original = $text

    # 1. NewAppContext(&pgxpool.Pool{}, ...) → NewAppContext(appx.MustNewPool(&pgxpool.Pool{}), ...)
    # 仅当 NewAppContext( 后跟 &pgxpool.Pool{} 时
    $text = $text -replace 'NewAppContext\(\&pgxpool\.Pool\{\}(,\s*nil\)\)', 'NewAppContext(appx.MustNewPool(&pgxpool.Pool{}), nil)'
    $text = $text -replace 'NewAppContext\((\s*pool\s*,)', 'NewAppContext(appx.MustNewPool($1),'
    $text = $text -replace 'NewAppContext\(\&pgxpool\.Pool\{\}(,\s*&)', 'NewAppContext(appx.MustNewPool(&pgxpool.Pool{}), &'
    $text = $text -replace 'NewAppContext\(\&pgxpool\.Pool\{\}\)', 'NewAppContext(appx.MustNewPool(&pgxpool.Pool{}))'
    $text = $text -replace 'NewAppContext\(\&pgxpool\.Pool\{\},\s*nil,\s*&config\.Config\{\}\)', 'NewAppContext(appx.MustNewPool(&pgxpool.Pool{}), nil, &config.Config{})'

    # 2. ctx.DB() 改为 ctx.DB().Raw()（不在 NewAppContext 内部）
    # 注意避开 NewAppContext 内部
    $text = $text -replace 'ctx\.DB\(\)(?![\.\w])', 'ctx.DB().Raw()'
    
    # 3. 也要加 appx import（如果还没有）
    if ($text -notmatch 'import\s+\([^)]*"gx1727\.com/xin/framework/pkg/appx"') {
        $text = $text -replace '("github\.com/jackc/pgx/v5/pgxpool")', '$1' + "`r`n`t" + '"gx1727.com/xin/framework/pkg/appx"'
    }

    if ($text -ne $original) {
        Set-Content -Path $path -Value $text -Encoding UTF8 -NoNewline
        Write-Host "patched: $path"
    } else {
        Write-Host "skipped (no change): $path"
    }
}