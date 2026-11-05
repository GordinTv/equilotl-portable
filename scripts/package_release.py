"""Package only audited source and explicitly named distribution files."""
import argparse
import hashlib
import json
import pathlib
import shutil
import subprocess
import tempfile
import zipfile
from audit_publication import ROOT, audit_current

parser = argparse.ArgumentParser()
parser.add_argument("--binaries", default="build/release-candidate")
parser.add_argument("--output", default="dist")
args = parser.parse_args()
files, findings = audit_current()
if findings:
    raise SystemExit("Publication audit failed; run scripts/audit_publication.py for redacted locations")
if subprocess.check_output(["git", "status", "--porcelain"], cwd=ROOT, text=True).strip():
    raise SystemExit("Commit audited source before packaging")
version = (ROOT / "VERSION").read_text().strip()
destination = (ROOT / args.output).resolve()
if not destination.is_relative_to(ROOT / "dist"):
    raise SystemExit("Distribution must be inside project/dist")
package = destination / ("Equilotl-Portable-" + version)
if package.exists():
    raise SystemExit("Distribution already exists; choose a new dist subdirectory for a fresh package")
package.mkdir(parents=True)
binroot = (ROOT / args.binaries).resolve()
metadata = json.loads((binroot / "BUILD-INFO.json").read_text(encoding="utf-8-sig"))
if metadata["version"] != version:
    raise SystemExit("Binary version differs from VERSION")
if metadata["commit"] != subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip():
    raise SystemExit("Build metadata does not match the current commit")
for name in ("Equilotl-Portable-windows-amd64.exe", "Equilotl-Portable-CLI-windows-amd64.exe", "BUILD-INFO.json"):
    shutil.copyfile(binroot / name, package / name)
expected = dict((name, sha) for sha, name in (line.split("  ", 1) for line in (binroot / "SHA256SUMS.txt").read_text(encoding="utf-8-sig").splitlines()))
for name in ("Equilotl-Portable-windows-amd64.exe", "Equilotl-Portable-CLI-windows-amd64.exe"):
    if hashlib.sha256((package / name).read_bytes()).hexdigest() != expected.get(name):
        raise SystemExit("Build artifact checksum mismatch")
for name in ("LICENSE", "NOTICE.md", "THIRD_PARTY_NOTICES.md", "README.md", "README.pt-BR.md", "CHANGELOG.md", "SECURITY.md"):
    shutil.copyfile(ROOT / name, package / name)
for name in ("docs", "third_party"):
    shutil.copytree(ROOT / name, package / name)
source = destination / ("Equilotl-Portable-" + version + "-source.zip")
# Include non-system linked dependency source, not only links to module servers.
# A dedicated ignored staging directory avoids touching any user vendor tree.
stage = pathlib.Path(tempfile.mkdtemp(prefix="ccs-", dir=ROOT / "build"))
vendored = stage / "vendor"
subprocess.run(["go", "mod", "vendor", "-o", str(vendored)], cwd=ROOT, check=True)
with zipfile.ZipFile(source, "w", zipfile.ZIP_DEFLATED) as archive:
    for name in files:
        archive.write(ROOT / name, "Equilotl-Portable/" + name)
    for p in sorted(vendored.rglob("*")):
        if p.is_file():
            archive.write(p, "Equilotl-Portable/vendor/" + p.relative_to(vendored).as_posix())
# Copy the exact corresponding source beside the binaries, not an upstream link.
shutil.copyfile(source, package / source.name)
manifest = {p.relative_to(package).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(package.rglob("*")) if p.is_file()}
(package / "SHA256SUMS.txt").write_text("".join(f"{sha}  {name}\n" for name, sha in manifest.items()), encoding="utf-8")
archive_path = destination / ("Equilotl-Portable-" + version + "-windows-amd64.zip")
with zipfile.ZipFile(archive_path, "w", zipfile.ZIP_DEFLATED) as archive:
    for p in sorted(package.rglob("*")):
        if p.is_file(): archive.write(p, package.name + "/" + p.relative_to(package).as_posix())
archives = (source, archive_path)
(destination / "SHA256SUMS.txt").write_text("".join(hashlib.sha256(p.read_bytes()).hexdigest() + "  " + p.name + "\n" for p in archives), encoding="utf-8")
print("Prepared clean distribution:", package.relative_to(ROOT).as_posix())
