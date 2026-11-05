# Installation and recovery

## Download and requirements

Use Windows x64 with a separately installed Discord Lite host. The recorded Lite version is PTB 1.0.1160. Preserve its distributor's pinned-version/updater setup. Download the Windows ZIP and checksum file from [this fork's releases](https://github.com/GordinTv/equilotl-portable/releases). Extract outside Discord's application/account directories. Check the entire published filename/hash:

```powershell
Get-FileHash .\Equilotl-Portable-0.1.0-rc.1-windows-amd64.zip -Algorithm SHA256
```

## Install

1. Quit Lite through its menu/notification-area icon. A closed window can leave background processes.
2. Open `Equilotl-Portable-windows-amd64.exe` as your normal user.
3. Select the displayed Lite channel/version and review its exact path/state.
4. For a custom location, choose `DC-Antilag`, its `app` base, or `app-VERSION`. A random directory named `app` is not sufficient.
5. Click **Install**, wait for success and start Lite using its usual launcher.
6. Sign in yourself, open settings and locate Equicord. Test a small optional plugin and persistence after restart. The installer never asks for credentials/tokens.

The existing `resources/app/package.json` entry changes; the original bootstrap stays in place. Executables, DLLs, modules and account files are not patched.

## CLI, repair and uninstall

Run from the extracted package:

```powershell
$hostPath = "$env:APPDATA\DC-Antilag\app\app-1.0.1160"
.\Equilotl-Portable-CLI-windows-amd64.exe --list
.\Equilotl-Portable-CLI-windows-amd64.exe --diagnose-processes --location $hostPath
.\Equilotl-Portable-CLI-windows-amd64.exe --install --location $hostPath
.\Equilotl-Portable-CLI-windows-amd64.exe --repair --location $hostPath
.\Equilotl-Portable-CLI-windows-amd64.exe --uninstall --location $hostPath
```

Diagnosis is read-only and does not terminate processes. Confirmed blockers include name, PID, path and reason; unavailable paths are inconclusive. Installation still requires the independent executable-access probe.

Fully quit Lite before Repair/Uninstall. **Reinstall / Repair** restores/reapplies the loader and downloads the latest Equicord. Downloads are staged before replacement; the prior build is retained in `equicord.asar.previous`. Installer upgrades are manual from this fork, separately from Equicord Repair.

**Uninstall** restores original manifest bytes/mode/modification time, preserves the DACL and removes verified installer-owned loader/state. It does not delete account, plugin settings or themes. Select each patched version separately when several versions exist.

## Recovery

Keep the entire `resources/app/.equilotl-portable` directory: original/patched manifests, loader/config and hash metadata.

1. Quit the host and run diagnosis.
2. Retry Uninstall to restore, or Repair to restore/reapply.
3. If integrity checks report external changes, stop forced edits. Unknown content is deliberately not overwritten.
4. Preserve a private copy of the application version and recovery directory for review. Do not delete state or copy `original.json` onto arbitrary paths to bypass checks.
5. If recovery cannot be completed, restore your own verified pre-install application copy or reinstall the application from its distributor while preserving account data separately. Never upload that application copy to an issue.

Recovery from physical storage corruption/power loss is not guaranteed. Unknown temporary files are retained conservatively. See [troubleshooting](TROUBLESHOOTING.md).
