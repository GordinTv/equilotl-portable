# Contributing

Start with an issue describing the behavior, host/fork versions and expected outcome. Use private vulnerability reporting for security defects. Do not submit account data, tokens, complete Discord copies, proprietary binaries or unredacted logs.

Keep the root Go module layout and separate portable/ASAR strategies. Preserve upstream copyright and dependency notices. A portable change must retain original bootstrap code, verified recovery and active-process protection. Do not bypass checks merely to hide an error.

Use the controlled toolchain in docs/BUILDING.md. Run formatting, both UI variants, race, vet, the publication policy and vulnerability checks. Add focused cases for safety regressions and document any manual-only behavior. CI uses fixtures/mocks; real Discord integration is opt-in/local-only and must use a new isolated copy.

Submit a focused PR explaining trigger, resulting behavior, preserved invariants, tests/evidence and limitations. Use small commits with clear subjects. Do not edit upstream history or force-push main. Link upstream issues when relevant without presenting this project as official.

Changes are contributed under the existing project GPL terms; no copyright assignment or invented CLA is required. Do not introduce dependencies without notice/license and security review. Update the locked graph and collected notices together. Release tags are a maintainer action after review, not part of ordinary PRs.
