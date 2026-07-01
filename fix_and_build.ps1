# fix_and_build.ps1 — 一键"扫坏文件→重写 UTF-8→编译验证"
#
# 用法（在仓库根目录 PowerShell 下）：
#   .\fix_and_build.ps1                 # 默认行为
#   .\fix_and_build.ps1 -CheckOnly      # 只扫不修
#   .\fix_and_build.ps1 -SkipUI         # 跳过 npm typecheck（前端有 node_modules 时省时）
#   .\fix_and_build.ps1 -GoOnly         # 只跑后端 build
#
# 设计目的：AGENTS.md §1"UTF-8 无 BOM"违反时，PowerShell GBK 写入会让 Go/TS 编译报
# illegal UTF-8 encoding。本脚本：
#   1) 跑 fix_encoding.py（GBK/GB18030/Big5 → UTF-8 回环校验转码）
#   2) 跑 strip_bom.py（剥 UTF-8 BOM）
#   3) 跑 go build ./...（后端编译）—— Go 编译器本身就是最权威的 UTF-8 校验器
#   4) 可选：跑 npm run typecheck（前端 TS 编译）
# 任何一步失败立刻退出，给出下一处坏文件路径。

[CmdletBinding()]
param(
    [switch]$CheckOnly,
    [switch]$SkipUI,
    [switch]$GoOnly
)

$ErrorActionPreference = "Stop"
$root = Resolve-Path .
$server = Join-Path $root "server"
$ui = Join-Path $root "UI"

function Step([string]$title) {
    Write-Host ""
    Write-Host "===== $title =====" -ForegroundColor Cyan
}

function Run([scriptblock]$cmd) {
    & $cmd
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[FAIL] 上一步退出码 $LASTEXITCODE — 终止。请把上面 go/tsc 报错的第一个非法文件路径发给我。" -ForegroundColor Red
        exit $LASTEXITCODE
    }
}

Step "1/4  scan+fix encoding (GBK/GB18030/Big5 → UTF-8)"
if ($CheckOnly) {
    Run { python (Join-Path $server "scripts/fix_encoding.py") --check (Join-Path $root "") }
} else {
    # 即使 --check 也跑一次让脚本自己报坏文件清单；如果真有坏文件但又不在
    # CANDIDATE_ENCODINGS 里，就跳到 fix-encoding 让它写回环后再报告。
    Run { python (Join-Path $server "scripts/fix_encoding.py") (Join-Path $root "") }
}

Step "2/4  strip UTF-8 BOM"
Run { python (Join-Path $server "scripts/strip_bom.py") (Join-Path $root "") }

Step "3/4  go build ./... (server)"
Push-Location $server
try {
    if ($CheckOnly) {
        Run { go build ./... }
    } else {
        Run { go build ./... }
    }
} finally {
    Pop-Location
}

if ($GoOnly) {
    Write-Host ""
    Write-Host "[done] GoOnly 指定，跳过 UI" -ForegroundColor Green
    exit 0
}

if (Test-Path (Join-Path $ui "package.json")) {
    Step "4/4  npm run typecheck (UI)"
    Push-Location $ui
    try {
        if (Test-Path "node_modules") {
            Run { npm run typecheck }
        } else {
            Write-Host "[skip] node_modules 不存在，先跑 npm install" -ForegroundColor Yellow
        }
    } finally {
        Pop-Location
    }
} else {
    Write-Host "[skip] 未找到 UI/package.json" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "[done] 全部通过" -ForegroundColor Green
