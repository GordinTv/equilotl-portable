param([string]$Version = '', [string]$Commit = '', [string]$OutputDirectory = '')
$ErrorActionPreference = 'Stop'
$project = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
Push-Location $project
$saved = @{}; foreach($name in @('CGO_ENABLED','GOOS','GOARCH')) { $saved[$name]=[Environment]::GetEnvironmentVariable($name,'Process') }
try {
    $config=Get-Content scripts/toolchain.json -Raw|ConvertFrom-Json
    if ((& go env GOVERSION) -ne ('go'+$config.go)) { throw "Distribution builds require Go $($config.go)" }
    if (-not (Get-Command gcc.exe -ErrorAction SilentlyContinue)) { throw 'Run scripts/Setup-Toolchain.ps1 first' }
    if ((& gcc.exe -dumpfullversion).Trim() -ne $config.gcc) { throw "Distribution builds require GCC $($config.gcc) from the pinned toolchain" }
    if (-not $Version) { $Version=(Get-Content VERSION -Raw).Trim() }
    $Version=$Version.TrimStart('v')
    if ($Version -notmatch '^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$') { throw 'Invalid semantic version' }
    if (-not $Commit) { $Commit=(& git.exe rev-parse HEAD).Trim();if($LASTEXITCODE){throw 'Git commit required'} }
    if ($Commit -notmatch '^[0-9a-f]{40}$') { throw 'Commit must be a full Git SHA' }
    $dirty=@(& git.exe status --porcelain --untracked-files=normal)
    if($dirty){throw 'Commit reviewed source before building a release candidate'}
    if (-not $OutputDirectory) { $OutputDirectory=Join-Path $project 'build\release-candidate' }
    $out=[IO.Path]::GetFullPath($OutputDirectory);New-Item -ItemType Directory -Force $out|Out-Null
    $env:GOOS='windows';$env:GOARCH='amd64';$env:CGO_ENABLED='1'
    $flags="-s -w -extldflags=-static -X equilotl/buildinfo.InstallerTag=$Version -X equilotl/buildinfo.InstallerGitHash=$Commit"
    & go build -mod=readonly -trimpath -buildvcs=false -tags cli -ldflags $flags -o (Join-Path $out 'Equilotl-Portable-CLI-windows-amd64.exe') .
    if($LASTEXITCODE){throw 'CLI compilation failed'}
    & go build -mod=readonly -trimpath -buildvcs=false -ldflags ("-H=windowsgui "+$flags) -o (Join-Path $out 'Equilotl-Portable-windows-amd64.exe') .
    if($LASTEXITCODE){throw 'GUI compilation failed'}
    @{version=$Version;commit=$Commit;go=(& go env GOVERSION);target='windows/amd64';cgo=$true;gcc=(& gcc.exe --version|Select-Object -First 1);flags='-trimpath -buildvcs=false; static external runtime';installer_self_update='disabled'}|ConvertTo-Json|Set-Content (Join-Path $out 'BUILD-INFO.json')
    Get-FileHash (Join-Path $out '*.exe')|ForEach-Object {$_.Hash.ToLower()+'  '+[IO.Path]::GetFileName($_.Path)}|Set-Content (Join-Path $out 'SHA256SUMS.txt')
} finally {foreach($name in $saved.Keys){[Environment]::SetEnvironmentVariable($name,$saved[$name],'Process')};Pop-Location}
