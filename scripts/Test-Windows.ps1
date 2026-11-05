param([string]$LogDirectory = 'build/validation')
$ErrorActionPreference='Stop'
Push-Location (Join-Path $PSScriptRoot '..')
$oldCGO=$env:CGO_ENABLED
try {
    $config=Get-Content scripts/toolchain.json -Raw|ConvertFrom-Json
    if ((& node --version) -ne ('v'+$config.node)) { throw "Tests require Node $($config.node); the loader contract test must not be skipped" }
    New-Item -ItemType Directory -Force $LogDirectory|Out-Null
    $files=@(git.exe ls-files --cached --others --exclude-standard '*.go');$unformatted=@(& gofmt -l $files)
    if($unformatted){throw ('Run gofmt on: '+($unformatted -join ', '))}
    & go mod verify;if($LASTEXITCODE){throw 'Module integrity check failed'}
    $env:CGO_ENABLED='1'
    foreach($variant in @('cli','gui','race')) {
        $goArgs=@('test','-count=1','-v');if($variant -eq 'cli'){$goArgs+=@('-tags','cli')};if($variant -eq 'race'){$goArgs+=@('-race','-tags','cli')};$goArgs+=@('./...')
        & go @goArgs *> (Join-Path $LogDirectory "tests-$variant.txt")
        if($LASTEXITCODE){
            # CI runs only synthetic fixtures; surface failures instead of losing runner-only logs.
            if($env:GITHUB_ACTIONS -eq 'true') { Get-Content (Join-Path $LogDirectory "tests-$variant.txt") | Write-Output }
            throw "Tests failed: $variant (inspect local logs)"
        }
        $checks = @(Select-String -Path (Join-Path $LogDirectory "tests-$variant.txt") -Pattern '^\s*--- PASS:').Count
        Write-Output "${variant}: $checks passing checks (including subtests)"
    }
    & go vet -tags cli ./...;if($LASTEXITCODE){throw 'CLI vet failed'}
    & go vet ./...;if($LASTEXITCODE){throw 'GUI vet failed'}
    Write-Output 'Formatting, module integrity, CLI/GUI tests, race and vet passed. No real Discord is launched.'
} finally {$env:CGO_ENABLED=$oldCGO;Pop-Location}
