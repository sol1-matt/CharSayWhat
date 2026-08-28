$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

& "$PSScriptRoot\make_ico.ps1" | Out-Null

$rsrc = (Get-Command rsrc -ErrorAction SilentlyContinue).Source
if (-not $rsrc) { $rsrc = 'D:\repos\bin\rsrc.exe' }
& $rsrc -ico charsaywhat.ico -arch amd64 -o rsrc_windows_amd64.syso

go build -ldflags="-H windowsgui" -o charsaywhat.exe .
