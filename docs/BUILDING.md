# Controlled Windows builds

Install Git, Go **1.26.10**, PowerShell 7 and Python 3.11+. Node 24 supports the loader-contract test. No real Discord is required for CI. Module name/import layout remain unchanged.

```powershell
git clone https://github.com/GordinTv/equilotl-portable.git
cd equilotl-portable
.\scripts\Setup-Toolchain.ps1
go mod download
python scripts/collect_licenses.py
.\scripts\Test-Windows.ps1
.\scripts\Build-Windows.ps1
python scripts/package_release.py
```

Setup verifies the fixed SHA-256 of official w64devkit **2.11.0** and adds GCC to the session. It does not install globally or change security settings. Go comes from its official installer or `actions/setup-go`.

The build script enforces controlled Go version, semantic version and full Git commit, targets Windows/amd64, uses the locked module graph, `-mod=readonly`, `-trimpath`, `-buildvcs=false`, and static external runtime linkage. It emits both binaries, build metadata and hashes under ignored `build/release-candidate`. VERSION defaults to `VERSION`; commit defaults to HEAD. The original Makefile remains for legacy development; this PowerShell script is the release path.

The packager copies only named binaries, docs/notices and audited corresponding source into a fresh ignored `dist` directory. The source ZIP includes a fresh `go mod vendor` snapshot of non-system linked dependencies, including C/C++ binding sources; it does not only link to third-party source servers. Staging stays in a dedicated ignored build directory. It refuses to overwrite an existing distribution. For another run:

```powershell
.\scripts\Build-Windows.ps1 -Version 0.1.0-rc.1 -OutputDirectory build/another-build
python scripts/package_release.py --binaries build/another-build --output dist/another-run
```

Release builds require a clean committed tree. Keep corresponding source and required notices with binary distribution. `third_party/modules.json` inventories locked module notices. Historical testing used Go 1.26.2; distribution moved to 1.26.10 after reachable standard-library vulnerabilities were identified.

Controlled inputs are not a universal byte-for-byte reproducibility claim. Runner images, archive timestamps and linker/toolchain behavior may differ. Windows binaries are unsigned unless a release explicitly states otherwise.
