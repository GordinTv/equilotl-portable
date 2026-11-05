# Release procedure

Initial proposed version: **0.1.0-rc.1**, independently versioned from upstream 2.3.0. VERSION omits the `v` prefix; tags use `v0.1.0-rc.1`. A release candidate communicates remaining manual acceptance limits.

1. Review source/history audits, licenses/artwork, corresponding source and the manual test matrix.
2. Enable repository Actions and Private Vulnerability Reporting. Configure the `release` environment with required reviewers. Keep workflow token permissions minimal.
3. Complete publication.json truthfully. A true authorization flag records an explicit owner decision, not an automated inference. Enable repository variable `ENABLE_RELEASE_PUBLISHING=true` only when release publication is approved/configured.
4. Commit the reviewed version and ensure CI passes on that commit. Do not publish a dirty local tree.
5. Push only the intended main branch and the new reviewed semantic version tag. Never mirror/push all inherited upstream tags: historical workflows belong to upstream commits.
6. The tag workflow validates the version, reruns Windows tests/security/builds, packages corresponding source/notices, and emits hashes and notes. Ordinary commits do not publish releases.
7. The publishing job additionally requires the opt-in variable, publication.json gates and protected environment. It creates a prerelease for an `-rc` tag. If gates are absent, artifacts can be reviewed without public release publication.
8. Verify every published asset/hash/source archive and the actual download links. Record the hosted run URL and known limitations. Do not describe artifacts as available until verified.

Alternatively use **Actions → Release → Run workflow**, enter the reviewed tag and request publication. This explicit manual request does not require the automatic-tag opt-in variable, but still requires the file gates and protected environment approval. It performs the same checks and creates the tag at that run's exact commit only inside the approved publishing job. With publication unchecked or gates disabled, it only prepares review artifacts. This is an explicit release action, not an ordinary-commit trigger.

Distribute the named GUI/CLI executables, Windows ZIP, source ZIP, BUILD-INFO.json, SHA256SUMS and required license/notices/docs. Never upload raw local validation logs, private paths, profiles, integration copies or unrelated old builds. Source publication and binary-release approval are distinct; user authorization to publish code alone does not silently enable the release gate.

GitHub settings cannot be proven by editing this file: enable them in the repository and verify them. The local prepared distribution is suitable for review but does not imply that hosted CI or a public release has run. The Actions workflow deliberately has no upstream Winget submission or inherited signing secrets.
