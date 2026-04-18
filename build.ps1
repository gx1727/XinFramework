# build.ps1
$OutputName = "xin-server.exe"
$BuildPath = ".\cmd\server\main.go"

Write-Host "Building $OutputName..." -ForegroundColor Green

go build -ldflags="-s -w" -o $OutputName $BuildPath

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host "Output: $OutputName" -ForegroundColor Cyan

    $FileSize = (Get-Item $OutputName).Length / 1MB
    Write-Host "Size: $([math]::Round($FileSize, 2)) MB" -ForegroundColor Cyan
} else {
    Write-Host "Build failed!" -ForegroundColor Red
}
