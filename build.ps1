# build.ps1
$OutputName = "xin-server.exe"
$BuildPath = ".\cmd\server\main.go"
$OutDir = ".\out"

# Create output directory
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
    
    # Copy config files to output directory
    Write-Host "Copying configuration files..." -ForegroundColor Green
    if (Test-Path ".\config") {
        if (!(Test-Path "$OutDir\config")) {
            New-Item -ItemType Directory -Path "$OutDir\config" | Out-Null
        }
        Copy-Item -Path ".\config\*" -Destination "$OutDir\config\" -Recurse -Force
        Write-Host "Config files copied to $OutDir\config\" -ForegroundColor Cyan
    }
    
    # Copy migrations if exists
    if (Test-Path ".\migrations") {
        if (!(Test-Path "$OutDir\migrations")) {
            New-Item -ItemType Directory -Path "$OutDir\migrations" | Out-Null
        }
        Copy-Item -Path ".\migrations\*" -Destination "$OutDir\migrations\" -Recurse -Force
        Write-Host "Migration files copied to $OutDir\migrations\" -ForegroundColor Cyan
    }
    
    Write-Host "`nRelease package ready in '$OutDir' directory!" -ForegroundColor Green
} else {
    Write-Host "Build failed!" -ForegroundColor Red
}
