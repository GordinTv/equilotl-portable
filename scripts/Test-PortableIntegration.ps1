param(
    [string]$SourceRoot = (Join-Path $env:APPDATA 'DC-Antilag'),
    [string]$Version = '1.0.1160',
    [string]$Installer = (Join-Path $PSScriptRoot '..\build\release-candidate\Equilotl-Portable-CLI-windows-amd64.exe'),
    [string]$EquicordAsar = (Join-Path $env:APPDATA 'Equicord\equicord.asar'),
    [switch]$Launch,
    [int]$Port = 19426
)
$ErrorActionPreference = 'Stop'
$project = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$sandbox = Join-Path $project ('build\integration-' + [guid]::NewGuid().ToString('N'))
$buildRoot = [IO.Path]::GetFullPath((Join-Path $project 'build')) + '\'
if (-not [IO.Path]::GetFullPath($sandbox).StartsWith($buildRoot,[StringComparison]::OrdinalIgnoreCase)) { throw 'Sandbox must remain inside build' }
if ($Version -notmatch '^\d+\.\d+\.\d+$') { throw 'Invalid version' }
$source = Join-Path $SourceRoot ('app\app-' + $Version)
if (Test-Path -LiteralPath (Join-Path $source 'resources\app\.equilotl-portable')) { throw 'Use an unpatched source copy for the integration baseline' }
$versionDir = Join-Path $sandbox ('DC-Antilag\app\app-' + $Version)
$evidence = Join-Path $sandbox 'evidence'
New-Item -ItemType Directory -Path $versionDir -Force | Out-Null
New-Item -ItemType Directory -Path $evidence -Force | Out-Null
# Copy executable/application files and public installation metadata only.
Get-ChildItem -LiteralPath $source -Force | ForEach-Object { Copy-Item -LiteralPath $_.FullName -Destination $versionDir -Recurse }
Copy-Item -LiteralPath (Join-Path $SourceRoot 'app\installer.db') -Destination (Join-Path $sandbox 'DC-Antilag\app\installer.db')
$data = Join-Path $sandbox 'DC-Antilag\data'
New-Item -ItemType Directory -Path $data | Out-Null
Copy-Item -LiteralPath (Join-Path $SourceRoot 'data\pinned_update.json') -Destination $data
[IO.File]::WriteAllText((Join-Path $data 'settings.json'),'{"SKIP_HOST_UPDATE":true,"USE_PINNED_UPDATE_MANIFEST":true,"OPTIN_OPTIONAL_UPDATES":true}')
$eq = Join-Path $sandbox 'Equicord'
New-Item -ItemType Directory -Path $eq | Out-Null
Copy-Item -LiteralPath $EquicordAsar -Destination (Join-Path $eq 'equicord.asar')
Get-FileHash -LiteralPath (Join-Path $eq 'equicord.asar') | Format-List | Out-File (Join-Path $evidence 'equicord-hash.txt')
$names = @('APPDATA','LOCALAPPDATA','EQUICORD_DEV_INSTALL','EQUICORD_DIRECTORY','EQUICORD_USER_DATA_DIR','DISCORD_USER_DATA_DIR')
$saved = @{}; foreach($name in $names) { $saved[$name] = [Environment]::GetEnvironmentVariable($name,'Process') }
function Invoke-Installer([string]$Action,[string]$Log) {
    & $Installer $Action --location $versionDir *> (Join-Path $evidence $Log)
    if ($LASTEXITCODE -ne 0) { throw "Installer failed: $Action; see $evidence\$Log" }
}
function Stop-IsolatedDiscord {
    $absoluteRoot = [IO.Path]::GetFullPath($sandbox) + '\'
    if (-not $absoluteRoot.StartsWith($buildRoot,[StringComparison]::OrdinalIgnoreCase)) { throw 'Unsafe process target' }
    Get-Process DiscordPTB -ErrorAction SilentlyContinue | Where-Object { $_.Path -and $_.Path.StartsWith($absoluteRoot,[StringComparison]::OrdinalIgnoreCase) } | Stop-Process -Force
}
try {
    $env:APPDATA = $sandbox
    $env:LOCALAPPDATA = Join-Path $sandbox 'NoOfficial'
    $env:EQUICORD_DEV_INSTALL = '1' # Test the supplied build without network downloads.
    $env:EQUICORD_DIRECTORY = Join-Path $eq 'equicord.asar'
    $env:EQUICORD_USER_DATA_DIR = $eq
    $env:DISCORD_USER_DATA_DIR = Join-Path $sandbox 'OfficialFallbackData'
    & node (Join-Path $PSScriptRoot 'compare-tree.cjs') $source $versionDir (Join-Path $evidence 'before.json')
    if ($LASTEXITCODE) { throw 'Initial copy differs' }
    Invoke-Installer '--install' 'install.txt'
    Invoke-Installer '--install' 'repeat.txt'
    Invoke-Installer '--repair' 'repair.txt'
    if ($Launch) {
        $listener = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback,$Port)
        try { $listener.Start() } finally { $listener.Stop() }
        Start-Process -FilePath (Join-Path $versionDir 'DiscordPTB.exe') -ArgumentList '--enable-logging',"--remote-debugging-port=$Port",'--no-first-run' -WindowStyle Hidden -RedirectStandardOutput (Join-Path $evidence 'stdout.txt') -RedirectStandardError (Join-Path $evidence 'stderr.txt') | Out-Null
        $ready = $false
        for($attempt=0;$attempt -lt 30;$attempt++) {
            Start-Sleep -Seconds 1
            $ErrorActionPreference = 'Continue'
            & node (Join-Path $PSScriptRoot 'probe-runtime.cjs') $Port (Join-Path $PSScriptRoot 'inspect-runtime.js') (Join-Path $evidence 'runtime.json') 2> (Join-Path $evidence 'probe-errors.txt')
            $ErrorActionPreference = 'Stop'
            if($LASTEXITCODE -eq 0) { $runtime=Get-Content (Join-Path $evidence 'runtime.json') -Raw | ConvertFrom-Json; if($runtime.equicord -eq 'object') { $ready=$true;break } }
        }
        if(-not $ready) { throw 'Equicord renderer did not become available' }
        $package = Join-Path $versionDir 'resources\app\package.json'
        $packageHash = (Get-FileHash -LiteralPath $package).Hash
        $ErrorActionPreference = 'Continue'
        & $Installer --install --location $versionDir *> (Join-Path $evidence 'running-refusal.txt')
        $refused = $LASTEXITCODE -ne 0
        $ErrorActionPreference = 'Stop'
        if(-not $refused -or (Get-FileHash -LiteralPath $package).Hash -ne $packageHash) { throw 'Running Discord was not safely refused' }
        & node (Join-Path $PSScriptRoot 'probe-runtime.cjs') $Port (Join-Path $PSScriptRoot 'test-runtime-settings.js') (Join-Path $evidence 'settings.json')
        if($LASTEXITCODE) { throw 'Runtime settings/plugins test failed' }
        Stop-IsolatedDiscord
    }
    Invoke-Installer '--uninstall' 'uninstall.txt'
    & node (Join-Path $PSScriptRoot 'compare-tree.cjs') $source $versionDir (Join-Path $evidence 'restoration.json')
    if($LASTEXITCODE) { throw 'Restoration differs from source' }
    Write-Output "PASS. Evidence: $evidence"
} finally {
    try { Stop-IsolatedDiscord } finally {
        foreach($name in $names) { [Environment]::SetEnvironmentVariable($name,$saved[$name],'Process') }
    }
}
