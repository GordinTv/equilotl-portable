# Testing and evidence

The portable/process revision recorded **88 passing checks including subtests** in both GUI and CLI variants. Publication work preserves them and adds upstream-self-update, DOS 8.3 alias and junction-ancestor regressions (91 checks per variant in the current local pass). These are table-inclusive counts, not 88 independent test functions. See the [sanitized validation record](validation/2026-11-05.md).

Run `scripts/Test-Windows.ps1` for formatting, module integrity, CLI/GUI tests, race detection and vet. Tests use temporary layouts, injected errors, local HTTP servers and harmless copies of the runner's own executable files. The Node loader test mocks Electron. CI must not start Discord or copy account data.

Coverage: official ASAR, portable/custom discovery, invalid layouts, repeated install, repair, restoration, interruption/rollback failures, NTFS ACLs/sharing, integrity/unknown changes, version changes, process exit/PID reuse, launchers/renamed helpers, other clients, casing/long paths and shared image aliases.

## Optional local integration

```powershell
.\scripts\Test-PortableIntegration.ps1 -Installer .\build\release-candidate\Equilotl-Portable-CLI-windows-amd64.exe -Launch
```

Requires an existing unpatched Lite host and local Equicord build. A new ignored `build` copy receives application files plus public installer/pinned-update metadata and new account-free settings. `DISCORD_USER_DATA_DIR` is set before launch to isolate official-updater fallback. Stopped processes are restricted to the copy's path; source files are not patched. Raw logs and proprietary copies remain private.

The script tests install/repeat/repair, startup before login, two plugins/native settings, refusal while running, uninstall and tree comparison. A recorded run restored 1,292 files without differences.

## Manual Windows checklist

Record version/date/result and sanitized reproduction. User reports remain labeled reported, separately from recorded checks.

| Check | Status before publication work |
| --- | --- |
| Install in isolated copy | Recorded CLI; successful real GUI install reported by maintainer |
| Open Lite with Equicord | Recorded before login; normal startup also reported |
| Login | Maintainer-reported; not automated/recorded by agent |
| Equicord settings after login | Maintainer-reported working |
| Enable a plugin after login | Maintainer-reported working; specific plugin/version not recorded |
| NoTrack/FixDiscordCss before login | Recorded renderer/bridge checks |
| Persist isolated setting | Recorded |
| Theme application/persistence | Pending |
| Voice/microphone/playback | Pending |
| Screen sharing | Pending |
| Restart with post-login settings | Pending explicit record |
| Repair/reopen | Recorded CLI; authenticated GUI acceptance pending |
| Uninstall/compare originals | Recorded 1,292-file equality |
| GUI process-detail scrolling | Compiled; visual acceptance pending |

Never automate authentication or publish private screenshots/session data. An early exploratory run preceded complete fallback isolation and allowed the native updater to open the default official PTB profile. Later scripts corrected this. No snapshot/absence-of-writes claim is made for that early profile; it is not distributed.
