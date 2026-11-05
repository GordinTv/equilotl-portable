# Architecture

| Area | Responsibility |
| --- | --- |
| `find_discord_*.go` | Platform discovery, host/version metadata |
| `portable.go` | Unpacked-host validation, verified transaction state, rollback/recovery |
| `portable_loader.cjs` | Embedded narrow Electron/Equicord path adapter |
| `portable_process_windows.go` | Process attribution/diagnosis and executable-access guards |
| `patcher.go`, `app_asar.go` | Strategy dispatch and original ASAR lifecycle |
| `github_downloader.go` | Equicord release metadata and staged build replacement |
| `gui.go`, `cli.go` | Interfaces sharing the same backend |
| `self_updater.go` | Disabled upstream installer-update API; Equicord Repair is separate |
| `scripts/`, workflows | Build, validation, publication policy, licenses and packaging |

Transaction: validate, acquire install lock, inspect processes, acquire executable guards, prepare/verify backup, write adapter, atomically replace manifest and release guards. Failure triggers restoration; a failed rollback preserves recovery data and reports both errors. Removal verifies integrity/ownership before restoring.

Account directories are outside the patch strategy. Equicord itself is separately downloaded and owns its settings. The fork adds no telemetry upload, private webhooks or stored Git credentials. Runtime network access remains for GitHub/Equicord release queries/downloads and OpenAsar for compatible official clients. Optional local runtime integration uses Discord network services in an isolated profile; CI never runs it.
