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

    if (Test-Path ".\migrations") {
        if (!(Test-Path "$OutDir\migrations")) {
            New-Item -ItemType Directory -Path "$OutDir\migrations" | Out-Null
        }
        Copy-Item -Path ".\migrations\*" -Destination "$OutDir\migrations\" -Recurse -Force
        Write-Host "Migration files copied to $OutDir\migrations\" -ForegroundColor Cyan
    }

    if (Test-Path ".\apps") {
        Get-ChildItem -Path ".\apps" -Directory | ForEach-Object {
            $appName = $_.Name

            if (Test-Path "$($_.FullName)\migrations") {
                $migrationsDest = "$OutDir\migrations\$appName"
                New-Item -ItemType Directory -Path $migrationsDest -Force | Out-Null
                Copy-Item "$($_.FullName)\migrations\*" "$migrationsDest\" -Recurse -Force
            }

            if (Test-Path "$($_.FullName)\config.yaml") {
                $cfgDest = "$OutDir\config\$appName"
                New-Item -ItemType Directory -Path $cfgDest -Force | Out-Null
                Copy-Item "$($_.FullName)\config.yaml" "$cfgDest\config.yaml" -Force
            }
        }
        Write-Host "App files copied" -ForegroundColor Cyan
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
