# Troubleshooting

| Symptom | Action |
| --- | --- |
| Invalid Location | Select the root/base/exact version and verify matching executable/build metadata and complete unpacked bootstrap. Do not create an arbitrary `app` directory. |
| Discord Lite is running / installation in use | Run diagnosis, inspect PID/path and quit the identified installation via its own menu/tray. Do not kill unrelated clients just by name. |
| Failed to patch this Install | Read the detailed reason: confirmed activity, inconclusive access, permission, download or integrity failure. Preserve recovery state. |
| Locked files / insufficient permissions | Close the relevant application and check ordinary directory/file access. The probe does not clear attributes/ACLs automatically. Do not disable security software. |
| Discord fails to start | Preserve a private copy, close the host, Uninstall on that exact version and check vanilla startup before Repair. If vanilla fails too, contact the Lite distributor. |
| Equicord settings absent | Confirm the launcher started the same patched version and the Equicord build exists. Fully quit, Repair and restart; a new host may be unpatched. |
| Incompatible host update | Re-detect the new version. Do not overwrite new code with an old bootstrap. Unsupported layouts should fail before mutation. |
| Backup/package integrity mismatch | Stop forced changes. Preserve all state and follow recovery; unknown content is intentionally not overwritten. |

```powershell
.\Equilotl-Portable-CLI-windows-amd64.exe --diagnose-processes --location "$env:APPDATA\DC-Antilag\app\app-1.0.1160"
```

An inconclusive query does not prove Lite is running. No attributed process does not bypass file safety either. Include fork version/commit, Windows architecture, host version, operation and sanitized error in bug reports. Replace local home names with `<user>`; process reports contain local paths. Do not attach account directories, tokens, cookies, passwords, raw session logs or complete Lite copies.

The tested Lite package logged a missing `discord_notifications` module while still starting. This project does not replace proprietary modules to hide that condition; notifications and voice are not certified by startup alone.
