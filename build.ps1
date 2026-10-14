# Builds the Windows binary into bin\charsaywhat.exe.
#   .\build.ps1            build (reuses the committed rsrc_windows_amd64.syso)
#   .\build.ps1 -Icon      also regenerate charsaywhat.ico and the .syso from charsaywhat.png
param([switch]$Icon)

$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

if ($Icon) {
    & "$PSScriptRoot\make_ico.ps1" | Out-Null

    $rsrc = (Get-Command rsrc -ErrorAction SilentlyContinue).Source
    if (-not $rsrc) { $rsrc = Join-Path (go env GOPATH) 'bin\rsrc.exe' }
    if (-not (Test-Path $rsrc)) {
        throw "rsrc not found. Install it with: go install github.com/akavel/rsrc@latest"
    }
    & $rsrc -ico charsaywhat.ico -arch amd64 -o rsrc_windows_amd64.syso
}

New-Item -ItemType Directory -Force -Path bin | Out-Null

$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
go build -trimpath -ldflags="-s -w" -o bin\charsaywhat.exe .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Output "Built bin\charsaywhat.exe"
