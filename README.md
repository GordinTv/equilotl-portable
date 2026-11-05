# Equilotl Portable — Discord Lite Support

[Português brasileiro](README.pt-BR.md) · [Installation](docs/INSTALLATION.md) · [Troubleshooting](docs/TROUBLESHOOTING.md)

[![CI](https://github.com/GordinTv/equilotl-portable/actions/workflows/ci.yml/badge.svg)](https://github.com/GordinTv/equilotl-portable/actions/workflows/ci.yml)
[![License: GPLv3](https://img.shields.io/badge/license-GPLv3-blue.svg)](LICENSE)

Independent community fork of [Equicord/Equilotl](https://github.com/Equicord/Equilotl), based on upstream **v2.3.0**, adding native support for unpacked Discord Lite portable installations.

**Unofficial:** not affiliated with Discord, officially with Equicord, or with Team AntiLag. Existing copyright, attribution and GPL terms are preserved. Client modifications may be subject to [Discord's Terms of Service](https://discord.com/terms).

## Features

- Automatic discovery of `%APPDATA%\DC-Antilag\app` and validated manual paths.
- Install, repair/update Equicord and uninstall through GUI or CLI.
- Dedicated `resources/app` strategy preserving the AntiLag bootstrap without patching executables, replacing DLLs or creating fake ASAR files.
- Verified backups, rollback and interrupted-install recovery.
- Process diagnostics by executable path, PID and liveness, including launchers and renamed helpers; unrelated installations are not attributed by name alone.
- Existing official Discord ASAR support retained; OpenAsar is deliberately unavailable for unpacked portable hosts.

## Download and quick start

Use [this fork's Releases](https://github.com/GordinTv/equilotl-portable/releases). Until an asset is published there, no public binary download is claimed. Upstream Equilotl downloads do not include these changes.

1. Obtain Discord Lite separately from a source you trust.
2. Download/extract the Windows x64 release ZIP and compare its SHA-256 with the published `SHA256SUMS.txt`.
3. Fully quit Lite, including notification-area/background processes.
4. Open `Equilotl-Portable-windows-amd64.exe`, select the intended Lite version/path and click **Install**.
5. Start Lite normally and check the Equicord settings section.

Use **Reinstall / Repair** to update Equicord or repair the loader; **Uninstall** restores the original manifest without deleting account or Equicord settings. [Detailed guide and recovery](docs/INSTALLATION.md).

```powershell
.\Equilotl-Portable-CLI-windows-amd64.exe --list
.\Equilotl-Portable-CLI-windows-amd64.exe --diagnose-processes --location "$env:APPDATA\DC-Antilag\app\app-1.0.1160"
```

The distribution contains GUI, CLI, documentation, required notices, checksums and corresponding source. It does not include Discord, Lite, Equicord, proprietary modules or account data. Binaries are unsigned unless a release states otherwise; checksums establish consistency, not publisher identity.

## Compatibility and limitations

Distribution target: **Windows x64**. Recorded runtime checks cover Lite PTB **1.0.1160** and official PTB **1.0.1230**. Fixtures cover other official channels. There is no universal compatibility guarantee.

Recorded integration reached login, loaded Equicord/two plugins, persisted an isolated setting and restored 1,292 files without differences. Post-login settings/plugins were subsequently **reported working by the maintainer**, separately from recorded automated evidence. Themes, voice, screen sharing and full manual acceptance remain pending explicit records. [Validation matrix](docs/TESTING.md).

Host/Equicord updates can require revalidation. Read-only files or insufficient permissions may prevent the executable safety probe without proving Lite is running. Do not disable security protections to bypass errors. Installer auto-update to upstream is disabled; Equicord Repair remains enabled. No private screenshots or raw account logs are published.

## Build, contribute and report

Use Go **1.26.10**, Windows x64, the checksum-pinned w64devkit in `scripts/toolchain.json`, Python 3.11+ for packaging and Node 24 for the loader test.

```powershell
.\scripts\Setup-Toolchain.ps1
.\scripts\Test-Windows.ps1
.\scripts\Build-Windows.ps1
python scripts/package_release.py
```

[Building](docs/BUILDING.md) · [Architecture](docs/ARCHITECTURE.md) · [Contributing](CONTRIBUTING.md) · [Changelog](CHANGELOG.md).

Report reproducible bugs in [Issues](https://github.com/GordinTv/equilotl-portable/issues), with sanitized errors only. Use [private vulnerability reporting](SECURITY.md) for security issues. Never attach tokens, cookies, account directories or complete Discord copies.

## License and credits

GNU GPLv3, retaining upstream terms and notices. See [LICENSE](LICENSE), [NOTICE](NOTICE.md) and [third-party notices](THIRD_PARTY_NOTICES.md). Equilotl derives from Vencord Installer; credits include Vendicated, Vencord contributors, thororen1234, Equicord contributors and dependency authors. No trademark rights or endorsement are claimed.
