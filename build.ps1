param(
    [string]$Version = "v0.1.0-alpha.1",
    [string]$OutputName = "yanp.exe"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$distDir = Join-Path $repoRoot "dist"
$binaryPath = Join-Path $distDir $OutputName
$zipName = "yanp-$Version-windows-amd64.zip"
$zipPath = Join-Path $distDir $zipName

New-Item -ItemType Directory -Force $distDir | Out-Null

Write-Host "Testing..."
go test -buildvcs=false ./...

Write-Host "Building $binaryPath ..."
go build -buildvcs=false -o $binaryPath .\cmd\yanp

if (Test-Path $zipPath) {
    Remove-Item -LiteralPath $zipPath -Force
}

Write-Host "Packaging $zipPath ..."
Compress-Archive -Path $binaryPath, README.md, docs\DEVLOG.md -DestinationPath $zipPath -Force

Write-Host ""
Write-Host "Build complete"
Write-Host "  Binary: $binaryPath"
Write-Host "  Archive: $zipPath"
