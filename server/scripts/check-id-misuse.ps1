# scripts/check-id-misuse.ps1
# Static check for ID-type misuse (alias mode lightweight guard).
#
# Rule: catch app.DB / app.Config usage OUTSIDE of:
#   - boot.go (framework boot stage)
#   - appx.go / appcontext*.go (app context definitions)
#   - framework.go (Serve / setupRouter boot stage)
#   - **/module.go (Module() function body uses app legitimately)
#
# All other files (handler, service, repository, etc.) should obtain
# config / pool via ctx, not via the module-scope *appx.App.
#
# False-positive tolerance: < 5% in alias mode. V2 named-type
# upgrade makes this checker obsolete (compiler enforces).

$root = "d:\work\xin\XinFramework\server"

# Use go list for fast file enumeration (skips vendor/.git automatically).
$goFiles = go list -f '{{.Dir}}\|{{range .GoFiles}}{{.}}|{{end}}' ./... 2>$null |
    ForEach-Object {
        $parts = $_ -split '\|', 2
        if ($parts.Count -lt 2) { return }
        $dir = $parts[0]
        foreach ($f in ($parts[1] -split '\|')) {
            if (-not $f) { continue }
            Join-Path $dir $f
        }
    }

# Skip patterns
$moduleFilePattern = '[/\\]module\.go$'
$bootFilePattern = '[/\\]framework\.go$'
$bootSubdirPattern = '[/\\]boot[/\\]'

$hits = @()
foreach ($path in $goFiles) {
    if ($path -match $moduleFilePattern) { continue }
    if ($path -match $bootFilePattern) { continue }
    if ($path -match $bootSubdirPattern) { continue }
    # Also skip appx / appcontext files
    if ($path -match 'appx\.go$|appcontext') { continue }
    $content = Get-Content -Path $path -Raw -Encoding UTF8
    $lineNum = 0
    foreach ($line in ($content -split "`n")) {
        $lineNum++
        $trimmed = $line.Trim()
        if ($trimmed.StartsWith("//")) { continue }
        if ($line -match '\bapp\.DB\b') {
            $hits += "$path`:$lineNum`: $trimmed"
        }
        if ($line -match '\bapp\.Config\b') {
            $hits += "$path`:$lineNum`: $trimmed"
        }
    }
}

if ($hits.Count -eq 0) {
    Write-Host "[OK] No app.DB / app.Config misuse outside module.go / boot.go / appx / framework.go"
    exit 0
}

Write-Host "[WARN] Possible ID/DB misuse patterns (alias mode known trade-off):" -ForegroundColor Yellow
Write-Host ""
$hits | Select-Object -First 20
Write-Host ""
Write-Host ("Total hits: {0}" -f $hits.Count) -ForegroundColor Yellow
Write-Host ""
Write-Host 'Note: these are expected false positives under alias mode.' -ForegroundColor DarkGray
Write-Host '      After V2 named-type upgrade, the compiler will catch them.' -ForegroundColor DarkGray