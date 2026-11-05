"""Fail-closed publication policy. Reports rule locations, never secret values."""
import argparse
import json
import pathlib
import re
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parents[1]
BAD_PARTS = {"build", "dist", ".git", "data", "node_modules", "dc-antilag", "logs", ".tools"}
BAD_SUFFIXES = {".exe", ".dll", ".node", ".asar", ".distro", ".db", ".log", ".zip", ".7z", ".old", ".bak", ".pfx", ".p12", ".key", ".pem"}
RULES = {
    "github-credential": re.compile(r"(?:gh[pousr]_[A-Za-z0-9]{36,}|github_pat_[A-Za-z0-9_]{50,})"),
    "private-key": re.compile(r"-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----"),
    "discord-webhook": re.compile(r"https://(?:\w+\.)?discord(?:app)?\.com/api/webhooks/\d{15,}/[A-Za-z0-9_-]{20,}"),
    "personal-windows-home": re.compile(r"(?i)[A-Z]:\\Users\\(?!User(?:\\|\b)|Public(?:\\|\b)|%USERNAME%(?:\\|\b)|<)[^\\\s\"']+"),
}

def git(*args, binary=False):
    return subprocess.check_output(["git", *args], cwd=ROOT, text=not binary)

def source_files():
    paths = git("ls-files", "-z", "--cached", "--others", "--exclude-standard", binary=True)
    return sorted({p.decode("utf-8") for p in paths.split(b"\0") if p and (ROOT / p.decode("utf-8")).is_file()})

def path_findings(name):
    p = pathlib.PurePosixPath(name.replace("\\", "/"))
    parts = {part.lower() for part in p.parts}
    if parts & BAD_PARTS or any(part.startswith(".equilotl-") for part in parts):
        return [{"file": name, "rule": "private-or-runtime-directory"}]
    if p.suffix.lower() in BAD_SUFFIXES or p.name.lower().startswith(".env") or ".sqlite" in p.name.lower():
        return [{"file": name, "rule": "non-source-or-sensitive-file"}]
    return []

def text_findings(name, data):
    if b"\0" in data:
        return []
    text = data.decode("utf-8", errors="replace")
    return [{"file": name, "line": text.count("\n", 0, m.start()) + 1, "rule": rule}
            for rule, regex in RULES.items() for m in regex.finditer(text)]

def audit_current():
    findings = []
    files = source_files()
    for name in files:
        if (ROOT / name).is_symlink() or not (ROOT / name).resolve().is_relative_to(ROOT):
            findings.append({"file": name, "rule": "source-link-or-escape"})
            continue
        findings += path_findings(name)
        findings += text_findings(name, (ROOT / name).read_bytes())
    return files, findings

def audit_history():
    # Read every reachable blob, including merge-only content, not just diffs.
    objects = git("rev-list", "--objects", "--all").splitlines()
    names = {}
    for line in objects:
        oid, _, name = line.partition(" ")
        if name:
            names.setdefault(oid, name)
    findings = []
    blobs = 0
    process = subprocess.Popen(["git", "cat-file", "--batch"], cwd=ROOT, stdin=subprocess.PIPE, stdout=subprocess.PIPE)
    try:
        for oid, name in names.items():
            process.stdin.write((oid + "\n").encode()); process.stdin.flush()
            header = process.stdout.readline().decode().split()
            size = int(header[2]); data = process.stdout.read(size); process.stdout.read(1)
            if header[1] != "blob":
                continue
            blobs += 1
            for item in path_findings(name) + text_findings(name, data):
                findings.append({**item, "blob": oid})
    finally:
        process.stdin.close(); process.wait()
    return blobs, findings

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--history", action="store_true")
    parser.add_argument("--output")
    parser.add_argument("--release-gate", action="store_true")
    args = parser.parse_args()
    files, findings = audit_current()
    report = {"source_files": len(files), "working_tree_findings": findings}
    if args.history:
        blobs, history = audit_history(); report.update(history_blobs=blobs, history_findings=history); findings += history
    if args.release_gate:
        config = json.loads((ROOT / "publication.json").read_text())
        for field in ("private_vulnerability_reporting_enabled", "release_environment_reviewers_configured", "redistribution_review_complete", "publication_authorized"):
            if config.get(field) is not True:
                findings.append({"file": "publication.json", "rule": "release-gate:" + field})
        owner = config.get("github_username", "")
        if not re.fullmatch(r"[A-Za-z0-9](?:[A-Za-z0-9-]{0,38})", owner):
            findings.append({"file": "publication.json", "rule": "release-gate:github_username"})
        report["release_gate_findings"] = [f for f in findings if f["rule"].startswith("release-gate:")]
    report["passed"] = not findings
    rendered = json.dumps(report, indent=2, ensure_ascii=False)
    if args.output:
        p = pathlib.Path(args.output); p.parent.mkdir(parents=True, exist_ok=True); p.write_text(rendered + "\n", encoding="utf-8")
    print(rendered)
    return 0 if report["passed"] else 1

if __name__ == "__main__":
    sys.exit(main())
