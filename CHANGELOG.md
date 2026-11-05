# Changelog

## Unreleased — planned 0.1.0-rc.1

First independent community release candidate, based on upstream Equilotl 2.3.0. A planned version is not evidence that a release has already been published.

### Added

- Discovery/manual selection of valid Team AntiLag portable hosts with unpacked resources/app.
- Dedicated Equicord loader preserving original bootstrap paths, without fake ASAR/executable/DLL replacement.
- Verified backup/rollback, interrupted-state recovery and external-change protection.
- Precise process attribution, exit/PID-reuse handling, launcher/helper coverage and read-only CLI diagnosis.
- Independent executable-access probes, including physical aliases and inconclusive process queries.
- Regression fixtures, controlled Windows builds, documentation, issue/PR templates and gated release workflow.

### Fixed

- Portable Invalid Location rejection and broad/name-based process attribution.
- Destructive/truncated Equicord downloads: downloads now stage/validate before replacement, retaining the previous build.
- Rollback rename order; stale official originals are retained in recovery archives.
- CLI error reporting and pre-existing vet formatting errors.

### Changed for distribution

- Explicit community identity and notices; stable public artifact names.
- Go 1.26.10 replaces the historically tested 1.26.2 toolchain after vulnerability analysis.
- Upstream installer self-update and download-and-run bootstraps are disabled to prevent replacing the fork. Equicord Repair remains enabled.

### Known limitations

Windows x64 is the distribution target. Themes, voice, screen sharing and complete authenticated GUI acceptance need recorded manual checks. Historical post-login settings/plugin success is maintainer-reported. Windows signing and universal bit-for-bit reproducibility are not claimed.
