# Publication preparation audit

## Origin and preservation

The supplied workspace contained no .git directory. Upstream Equicord/Equilotl refs and full available history were fetched read-only, with v2.3.0 as the functional base. Local portable changes were committed without dropping that ancestry. The public destination is the maintainer-created fork `GordinTv/equilotl-portable`. New upstream signing-workflow/support-link commits are not adopted as this fork's signing or support infrastructure; their history remains preserved.

No original repository history was rewritten. Do not push inherited tags wholesale. Local synthetic preparation-author metadata is explained in NOTICE.md; no private maintainer email is used as a contact address.

## Scope and identity

Go source remains at module root. Portable detection/patch/recovery/process code and the 88-check baseline were retained. Distribution changes are controlled toolchain/security updates, clear community labels/support links, disabled upstream installer replacement, documentation and maintenance automation. Original LICENSE and credits are retained; dependency/runtime notices are collected in third_party.

Inherited auto-release/Winget/signing workflows were removed from the active tree, with local copies retained only in ignored audit storage. No upstream signing secret, PAT, certificate or organization configuration is reused. Original artwork is unmodified, with no endorsement or trademark grant claimed; redistribution presentation remains a release review item.

## Exclusions and checks

Private build trees contain old binaries, complete test application copies, generated data, backups and raw logs. They were not deleted by assumption. They are excluded through Git policy and a fresh allowlisted packaging path. Neither those copies nor archived private reports belong in the source/release upload.

The publication checker inspects tracked/candidate paths and text, rejects private/runtime formats and source links, and checks every reachable historical blob, including merge-only content. Gitleaks supplements this with full-ref history detection. Findings are reported by rule/location, never secret value. Before initial publication the recorded history policy pass inspected 717 blobs without findings; final passes include new preparation commits as well.

Dependencies/tool versions are locked and audited. Go 1.26.2 had 19 reachable standard-library findings; Go 1.26.10 removed reachable findings in the recorded CLI/GUI analyses. Remaining unreachable reports and native-code limits are documented rather than presented as a blanket security guarantee.

## Distribution and remaining owner configuration

Clean output contains the two consistently named Windows executables, build metadata, hashes, guides/notices and corresponding source including a vendor snapshot. It excludes proprietary application binaries, DLLs, account/profile/session data and private logs. Unsigned builds and cross-platform/runtime limitations are disclosed.

Before public binary release, verify Actions is enabled, Private Vulnerability Reporting is enabled, the `release` environment has required reviewers, the publication flags accurately reflect approval/review, the publishing variable is enabled intentionally, and hosted checks pass for the chosen version/tag. Public source authorization does not by itself bypass these controls. See RELEASING.md for the exact owner actions and the planned 0.1.0-rc.1 release.
