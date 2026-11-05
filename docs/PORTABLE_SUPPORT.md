# Portable support design

Detection requires coherent numeric version/build metadata, a known channel/executable, a CommonJS package named `discord`, entry `app_bootstrap/index.js`, `common/paths.js` and `node_modules`. Relevant links/junctions and unexpected types are refused. The host is revalidated before mutation. A known damaged transaction may remain visible without authorizing unsafe changes.

Electron reads the package `main`. The strategy backs up original JSON bytes and points `main` to `.equilotl-loader.cjs`. AntiLag code stays in its original paths. The adapter translates only the virtual `_app.asar` manifest/entry requests expected by Equicord. It creates no ASAR file or directory. `app.setAppPath`, `require.main.filename` and module search paths are corrected; Equicord invokes the original entry once.

State format 1 records SHA-256 for original/patched manifests, loader/config, and original metadata. Writes are staged, flushed and verified; Windows replacement preserves destination security metadata. Rollback failures are reported alongside the initial error, retaining recovery state. Unknown changes/files are not discarded.

Processes are attributed to the exact selected version and its shared launcher, not names or the whole parent `app` directory. Liveness is checked around path queries and blockers are rechecked after enumeration to handle exit/PID reuse. Inconclusive observations remain separate.

The host, helper executables and shared launcher require exclusive-access probes. A read-only handle alone is insufficient for an already mapped image: the probe requests read/write access without writing bytes. Restrictive ACLs/read-only attributes may therefore cause refusal. Self PID/file identity are excluded; physical aliases remain protected.

The strategy does not modify account credentials, executables, DLLs, native modules or the host updater. New host versions require revalidation/reapplication. Official Discord keeps its existing ASAR strategy; OpenAsar is rejected for the unpacked portable layout. Linux/macOS source remains, but Windows x64 is the distributed/tested target.

References: [Equilotl upstream](https://github.com/Equicord/Equilotl), [Electron startup](https://github.com/electron/electron/blob/main/lib/browser/init.ts), [Equicord patcher](https://github.com/Equicord/Equicord/blob/main/src/main/patcher.ts). These contracts can evolve.
