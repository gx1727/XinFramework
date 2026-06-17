#requires -Version 5.1
<#
.SYNOPSIS
    XinFramework Phase 0 global-variable scan script.

.DESCRIPTION
    Scans every package-level global in server/ and locates definition
    sites plus all read/write call sites. Emits both a Markdown report
    and a machine-readable JSON. Output goes to doc/refactor/phase0/.

    The report is the input to Phase B (introduce AppContext). Each
    entry tells the refactor author which global to remove and which
    call sites must be rewritten.
#>

[CmdletBinding()]
param(
    [string] $Root,
    [string] $OutDir
)

if (-not $Root) {
    if ($PSScriptRoot) {
        $Root = Split-Path -Parent $PSScriptRoot
    } else {
        $Root = (Get-Location).Path
    }
}

$ErrorActionPreference = 'Continue'

if (-not $OutDir) {
    $OutDir = Join-Path $Root 'doc/refactor/phase0'
}
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

# ---------------------------------------------------------------------------
# 1. Known globals. Phase B will remove the first 12 and keep the last 4.
# ---------------------------------------------------------------------------
$globals = @(
    @{ Name = 'globalAccountFactory';      Category = 'remove'; Package = 'framework/pkg/auth';            File = 'framework/pkg/auth/registry.go' }
    @{ Name = 'globalAccountAuthFactory';  Category = 'remove'; Package = 'framework/pkg/auth';            File = 'framework/pkg/auth/registry.go' }
    @{ Name = 'globalFactory';             Category = 'remove'; Package = 'framework/pkg/tenant';          File = 'framework/pkg/tenant/registry.go' }
    @{ Name = 'globalUserFactory';         Category = 'remove'; Package = 'framework/pkg/rbac';            File = 'framework/pkg/rbac/user.go' }
    @{ Name = 'globalRoleFactory';         Category = 'remove'; Package = 'framework/pkg/rbac';            File = 'framework/pkg/rbac/role.go' }
    @{ Name = 'globalOrganizationFactory'; Category = 'remove'; Package = 'framework/pkg/rbac';            File = 'framework/pkg/rbac/organization.go' }
    @{ Name = 'globalPermissionFactory';   Category = 'remove'; Package = 'framework/pkg/rbac';            File = 'framework/pkg/rbac/permission.go' }
    @{ Name = 'globalProvider';            Category = 'remove'; Package = 'framework/pkg/extapi';          File = 'framework/pkg/extapi/provider.go' }
    @{ Name = 'global';                    Category = 'remove'; Package = 'framework/pkg/authz';           File = 'framework/pkg/authz/authz.go' }
    @{ Name = 'globalAuthorizationService';Category = 'remove'; Package = 'framework/internal/service';    File = 'framework/internal/service/authorization_service.go' }
    @{ Name = 'globalApp';                 Category = 'remove'; Package = 'framework/internal/core/boot';  File = 'framework/internal/core/boot/boot.go' }
    @{ Name = 'globalCache';               Category = 'remove'; Package = 'framework/pkg/dict';            File = 'framework/pkg/dict/dict.go' }
    @{ Name = 'Pool';                      Category = 'keep';   Package = 'framework/pkg/db';              File = 'framework/pkg/db/db.go' }
    @{ Name = 'Client';                    Category = 'keep';   Package = 'framework/pkg/cache';           File = 'framework/pkg/cache/cache.go' }
    @{ Name = 'cfg';                       Category = 'keep';   Package = 'framework/pkg/config';          File = 'framework/pkg/config/config.go' }
    @{ Name = 'defaultManager';            Category = 'keep';   Package = 'framework/pkg/session';         File = 'framework/pkg/session/session.go' }
)

# ---------------------------------------------------------------------------
# 2. Helpers
# ---------------------------------------------------------------------------
function Get-RelativePath([string] $abs) {
    if ($abs.StartsWith($Root)) {
        return $abs.Substring($Root.Length).TrimStart('\', '/') -replace '\\', '/'
    }
    return $abs
}

function Find-GlobalDef([string] $name, [string] $hintFile) {
    $pat = '^\s*var\s+' + [regex]::Escape($name) + '(\s|=)'
    $hintAbs = Join-Path $Root $hintFile
    if (Test-Path -LiteralPath $hintAbs) {
        $lines = Get-Content -LiteralPath $hintAbs
        for ($i = 0; $i -lt $lines.Count; $i++) {
            if ($lines[$i] -match $pat) {
                $hit = @{ File = $hintFile; Line = $i + 1; Match = $lines[$i].Trim() }
                return $hit
            }
        }
    }
    $allFiles = Get-ChildItem -Path $Root -Recurse -Filter '*.go' -ErrorAction SilentlyContinue
    $allHits = Select-String -Path $allFiles.FullName -Pattern $pat -ErrorAction SilentlyContinue
    if ($allHits) {
        $first = $allHits | Select-Object -First 1
        $hit = @{ File = (Get-RelativePath $first.Path); Line = $first.LineNumber; Match = $first.Line.Trim() }
        return $hit
    }
    return $null
}

function Find-Usages([string] $name) {
    $pat = '\b' + [regex]::Escape($name) + '\b'
    $files = Get-ChildItem -Path $Root -Recurse -Filter '*.go' -ErrorAction SilentlyContinue
    $hits = Select-String -Path $files.FullName -Pattern $pat -ErrorAction SilentlyContinue
    $out = New-Object System.Collections.Generic.List[object]
    if ($null -ne $hits) {
        foreach ($h in $hits) {
            $rel = Get-RelativePath $h.Path
            $out.Add([PSCustomObject]@{
                File    = $rel
                Line    = $h.LineNumber
                Snippet = $h.Line.Trim()
            })
        }
    }
    return ,$out
}

# ---------------------------------------------------------------------------
# 3. Scan
# ---------------------------------------------------------------------------
$report = New-Object System.Collections.Generic.List[object]
$totalUsages = 0

Write-Host ('Scanning ' + $globals.Count + ' known globals under ' + $Root)

foreach ($g in $globals) {
    Write-Host ('  - ' + $g.Name)
    $def = Find-GlobalDef -name $g.Name -hintFile $g.File
    $usages = Find-Usages -name $g.Name

    $writeCount = 0
    foreach ($u in $usages) {
        $pat2 = '=\s*' + [regex]::Escape($g.Name) + '(\b|$)'
        if ($u.Snippet -match $pat2) { $writeCount++ }
    }
    $readCount = $usages.Count - $writeCount

    $entry = [PSCustomObject]@{
        Name           = $g.Name
        Category       = $g.Category
        Package        = $g.Package
        DefinitionFile = if ($def) { $def.File } else { '(not found)' }
        DefinitionLine = if ($def) { $def.Line } else { 0 }
        Definition     = if ($def) { $def.Match } else { '' }
        TotalUsages    = $usages.Count
        WriteCount     = $writeCount
        ReadCount      = $readCount
        Usages         = $usages
    }
    $report.Add($entry)
    $totalUsages += $usages.Count
}

# ---------------------------------------------------------------------------
# 4. JSON
# ---------------------------------------------------------------------------
$jsonPath = Join-Path $OutDir 'globals.json'
$json = $report | ConvertTo-Json -Depth 6
[System.IO.File]::WriteAllText($jsonPath, $json, [System.Text.UTF8Encoding]::new($false))
Write-Host ('JSON: ' + $jsonPath)

# ---------------------------------------------------------------------------
# 5. Markdown (English labels, no PS5.1 parenthesized-string pitfall)
# ---------------------------------------------------------------------------
$mdPath = Join-Path $OutDir 'globals.md'
$md = New-Object System.Text.StringBuilder
$null = $md.AppendLine('# Phase 0 - Global Variable Inventory')
$null = $md.AppendLine('')
$null = $md.AppendLine('> Auto-generated. Re-run: `powershell scripts/phase0_scan.ps1`')
$null = $md.AppendLine('')
$null = $md.AppendLine('- Repo root: `' + $Root + '`')
$null = $md.AppendLine('- Tracked globals: **' + $report.Count + '**')
$null = $md.AppendLine('- Total usages: **' + $totalUsages + '**')
$null = $md.AppendLine('')

$null = $md.AppendLine('## 1. Cross-module globals (Phase B must remove)')
$null = $md.AppendLine('')
$null = $md.AppendLine('| Variable | Package | Defined at | Writes | Reads |')
$null = $md.AppendLine('|---|---|---|---:|---:|')
foreach ($r in ($report | Where-Object { $_.Category -eq 'remove' })) {
    $row = ('| `' + $r.Name + '` | `' + $r.Package + '` | ' + $r.DefinitionFile + ':' + $r.DefinitionLine + ' | ' + $r.WriteCount + ' | ' + $r.ReadCount + ' |')
    $null = $md.AppendLine($row)
}
$null = $md.AppendLine('')

$null = $md.AppendLine('## 2. Infrastructure globals (keep, surface through AppContext reader)')
$null = $md.AppendLine('')
$null = $md.AppendLine('| Variable | Package | Defined at | Reads |')
$null = $md.AppendLine('|---|---|---|---:|')
foreach ($r in ($report | Where-Object { $_.Category -eq 'keep' })) {
    $row = ('| `' + $r.Name + '` | `' + $r.Package + '` | ' + $r.DefinitionFile + ':' + $r.DefinitionLine + ' | ' + $r.ReadCount + ' |')
    $null = $md.AppendLine($row)
}
$null = $md.AppendLine('')

$null = $md.AppendLine('## 3. Detailed call sites')
$null = $md.AppendLine('')
foreach ($r in $report) {
    $null = $md.AppendLine('### `' + $r.Name + '`')
    $null = $md.AppendLine('')
    $null = $md.AppendLine('- Package: `' + $r.Package + '`')
    $null = $md.AppendLine('- Definition: `' + $r.DefinitionFile + ':' + $r.DefinitionLine + '`')
    $null = $md.AppendLine('- Usages: ' + $r.TotalUsages + ' total (write ' + $r.WriteCount + ' / read ' + $r.ReadCount + ')')
    $null = $md.AppendLine('')
    $null = $md.AppendLine('| File | Line | Snippet |')
    $null = $md.AppendLine('|---|---:|---|')
    foreach ($u in ($r.Usages | Sort-Object File, Line)) {
        $snippet = ($u.Snippet -replace '\|','\|' -replace '\s+',' ').Trim()
        if ($snippet.Length -gt 160) { $snippet = $snippet.Substring(0, 157) + '...' }
        $row = ('| `' + $u.File + '` | ' + $u.Line + ' | `' + $snippet + '` |')
        $null = $md.AppendLine($row)
    }
    $null = $md.AppendLine('')
}

[System.IO.File]::WriteAllText($mdPath, $md.ToString(), [System.Text.UTF8Encoding]::new($false))
Write-Host ('Markdown: ' + $mdPath)

# ---------------------------------------------------------------------------
# 6. Console summary
# ---------------------------------------------------------------------------
Write-Host ''
Write-Host '=== Summary ===' -ForegroundColor Yellow
$hdr = '{0,-32} {1,-50} {2,6} {3,6}' -f 'Variable','Package','Write','Read'
Write-Host $hdr
Write-Host ('-' * 100)
foreach ($r in $report) {
    $line = '{0,-32} {1,-50} {2,6} {3,6}' -f $r.Name, $r.Package, $r.WriteCount, $r.ReadCount
    Write-Host $line
}
Write-Host ''
$removeSum = ($report | Where-Object { $_.Category -eq 'remove' } | Measure-Object TotalUsages -Sum).Sum
$keepSum   = ($report | Where-Object { $_.Category -eq 'keep'   } | Measure-Object TotalUsages -Sum).Sum
Write-Host ('Remove-pool usages to rewrite: ' + $removeSum) -ForegroundColor Green
Write-Host ('Keep-pool usages to retarget:  ' + $keepSum)   -ForegroundColor Green
