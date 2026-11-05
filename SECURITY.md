# Security policy

Do not open a public issue containing a vulnerability, token, credential, local account profile or exploit against a third-party service.

Use GitHub **Private Vulnerability Reporting** for this repository: open **Security → Advisories → Report a vulnerability**, if that option is enabled. Repository maintainers must enable it before the first public binary release. No email address is invented here. If private reporting is unavailable, do not post sensitive details publicly; request that maintainers enable a private reporting channel using a non-sensitive issue.

Include affected fork/host versions, Windows architecture, impact, a minimal reproduction in a disposable fixture, and sanitized diagnostics. Do not attach a real Discord data directory or redistributable copy of proprietary application files. Share exploit material only through the private report.

Only the latest published release candidate/release is currently intended for maintenance. There is no promised response SLA while this community project is being established. Dependencies and the Go toolchain are checked with govulncheck; native code and unreachable reports still require human review. Secret scanning is performed on source/history and only approved files are packaged.

Installer self-update to upstream is disabled. Download manual installer upgrades from this fork, verify release checksums, and retain corresponding source/notices. Never disable endpoint protection or grant broad access simply to work around an installer failure. Read-only/access-denied probes are safety outcomes, not proof of a particular running process.
