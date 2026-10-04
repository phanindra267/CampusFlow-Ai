param(
    [Parameter(Mandatory = $true)][string]$Package,
    [string[]]$BuildArgs = @(),
    [string[]]$RunArgs = @(),
    [int]$Attempts = 10
)

# Builds a package and runs it, retrying while Windows Application Control
# blocks the freshly written executable. Each attempt writes to a unique path so
# a file that has already been blocked is not reused.
$ErrorActionPreference = 'Continue'
$name = [System.IO.Path]::GetFileNameWithoutExtension($Package)
$root = Split-Path -Parent $PSScriptRoot
$pkgPath = Join-Path $root $Package

for ($i = 1; $i -le $Attempts; $i++) {
    $exe = Join-Path $root ".bin\run-$name-$i.exe"
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $exe) | Out-Null
    if (Test-Path $exe) { Remove-Item -Force $exe }

    $build = & go build -o $exe $pkgPath 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Output "BUILD FAILED for $Package"
        $build | ForEach-Object { Write-Output $_ }
        exit 1
    }

    $blocked = $false
    $output = @()
    try {
        $output = & $exe @RunArgs 2>&1
        $blocked = (($output | Out-String) -match 'Application Control')
    } catch {
        $blocked = $true
    }

    if ($blocked) {
        Write-Output "attempt ${i}: blocked by Application Control"
        Start-Sleep -Seconds 2
        continue
    }

    $output | ForEach-Object { Write-Output $_ }
    exit $LASTEXITCODE
}

Write-Output "BLOCKED after $Attempts attempts"
exit 99