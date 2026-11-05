"""Validate package checksums, source paths and the explicit binary allowlist."""
import hashlib
import pathlib
import sys
import zipfile
from audit_publication import path_findings, text_findings

root = pathlib.Path(sys.argv[1]).resolve()
checksums = root / "SHA256SUMS.txt"
if not checksums.is_file():
    raise SystemExit("Missing distribution checksum file")
listed = set()
for line in checksums.read_text(encoding="utf-8-sig").splitlines():
    sha, name = line.split("  ", 1)
    p = root / name
    if p.parent != root or not p.is_file() or hashlib.sha256(p.read_bytes()).hexdigest() != sha:
        raise SystemExit("Archive hash/path mismatch")
    listed.add(name)
    with zipfile.ZipFile(p) as z:
        for entry in z.infolist():
            pure = pathlib.PurePosixPath(entry.filename)
            if pure.is_absolute() or ".." in pure.parts or len(pure.parts) < 2:
                raise SystemExit("Unsafe archive path")
            relative = pathlib.PurePosixPath(*pure.parts[1:]).as_posix()
            if name.endswith("-source.zip"):
                findings = path_findings(relative) + text_findings(relative, z.read(entry))
                if findings:
                    raise SystemExit("Source archive violates publication policy: " + str(findings[:5]))
            else:
                allowed = {"Equilotl-Portable-windows-amd64.exe", "Equilotl-Portable-CLI-windows-amd64.exe"}
                if pure.suffix.lower() == ".exe" and relative not in allowed:
                    raise SystemExit("Unexpected executable in distribution")
                if pure.suffix.lower() in {".dll", ".node", ".asar", ".db", ".log", ".distro"} or "data" in pure.parts:
                    raise SystemExit("Runtime/private content in distribution")
if len(listed) != 2:
    raise SystemExit("Expected corresponding source and Windows package")
print("Distribution archives and SHA-256 integrity verified")
