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

    if (Test-Path ".\framework\migrations") {
        if (!(Test-Path "$OutDir\framework\migrations")) {
            New-Item -ItemType Directory -Path "$OutDir\framework\migrations" | Out-Null
        }
        Copy-Item -Path ".\framework\migrations\*" -Destination "$OutDir\framework\migrations\" -Recurse -Force
        Write-Host "Migration files copied to $OutDir\framework\migrations\" -ForegroundColor Cyan
    }

    if (Test-Path ".\apps") {
        Get-ChildItem -Path ".\apps" -Directory | ForEach-Object {
            $appName = $_.Name

            if (Test-Path "$($_.FullName)\migrations") {
                $migrationsDest = "$OutDir\migrations\$appName"
                New-Item -ItemType Directory -Path $migrationsDest -Force | Out-Null
                Copy-Item "$($_.FullName)\migrations\*" "$migrationsDest\" -Recurse -Force
            }
        }
        Write-Host "App migration files copied" -ForegroundColor Cyan
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
