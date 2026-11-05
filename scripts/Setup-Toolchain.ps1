param([string]$CacheRoot = (Join-Path $env:TEMP 'equilotl-portable-toolchain'))
$ErrorActionPreference = 'Stop'
if (-not $IsWindows -and $env:OS -ne 'Windows_NT') { throw 'Windows toolchain required' }
$config = Get-Content (Join-Path $PSScriptRoot 'toolchain.json') -Raw | ConvertFrom-Json
$root = [IO.Path]::GetFullPath((Join-Path $CacheRoot $config.w64devkit.version))
New-Item -ItemType Directory -Force $root | Out-Null
$archive = Join-Path $root 'w64devkit.7z.exe'
if (-not (Test-Path -LiteralPath $archive)) { Invoke-WebRequest $config.w64devkit.url -OutFile $archive }
if ((Get-FileHash -LiteralPath $archive).Hash.ToLower() -ne $config.w64devkit.sha256) { throw 'Toolchain checksum mismatch; cached archive retained for inspection' }
$bin = Join-Path $root 'w64devkit\bin'
if (-not (Test-Path -LiteralPath (Join-Path $bin 'gcc.exe'))) {
    & $archive -y ('-o' + $root) | Out-Null
    if ($LASTEXITCODE) { throw 'Toolchain extraction failed' }
}
$env:PATH = "$bin;$env:PATH"
$env:CC = Join-Path $bin 'gcc.exe'
$env:CXX = Join-Path $bin 'g++.exe'
if ($env:GITHUB_PATH) { $bin | Out-File $env:GITHUB_PATH -Append -Encoding utf8 }
if ($env:GITHUB_ENV) {
    "CC=$env:CC" | Out-File $env:GITHUB_ENV -Append -Encoding utf8
    "CXX=$env:CXX" | Out-File $env:GITHUB_ENV -Append -Encoding utf8
}
Write-Output "Verified w64devkit $($config.w64devkit.version)"
