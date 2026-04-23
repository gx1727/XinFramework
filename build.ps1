$OutputName = "xin.exe"
$BuildPath = ".\cmd\xin"
$OutDir = ".\out"

if (!(Test-Path $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir | Out-Null
}

Write-Host "Building $OutputName..." -ForegroundColor Green

go build -ldflags="-s -w" -o "$OutDir\$OutputName" $BuildPath

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host "Output: $OutDir\$OutputName" -ForegroundColor Cyan

    $FileSize = (Get-Item "$OutDir\$OutputName").Length / 1MB
    Write-Host "Size: $([math]::Round($FileSize, 2)) MB" -ForegroundColor Cyan

    Write-Host "Copying configuration files..." -ForegroundColor Green
    if (Test-Path ".\config") {
        if (!(Test-Path "$OutDir\config")) {
            New-Item -ItemType Directory -Path "$OutDir\config" | Out-Null
        }
        Copy-Item -Path ".\config\*" -Destination "$OutDir\config\" -Recurse -Force
        Write-Host "Config files copied to $OutDir\config\" -ForegroundColor Cyan
    }

    Write-Host "Copying migration files..." -ForegroundColor Green
    if (Test-Path ".\migrations") {
        Copy-Item -Path ".\migrations" -Destination "$OutDir\migrations" -Recurse -Force
        Write-Host "Migration files copied to $OutDir\migrations\" -ForegroundColor Cyan
    }

    if (Test-Path ".\framework\.env.example") {
        Copy-Item ".\framework\.env.example" "$OutDir\.env.example" -Force
        Write-Host "Env example copied to $OutDir\.env.example" -ForegroundColor Cyan
    }

    Write-Host ""
    Write-Host "Release package ready in '$OutDir' directory!" -ForegroundColor Green
} else {
    Write-Host "Build failed!" -ForegroundColor Red
}
