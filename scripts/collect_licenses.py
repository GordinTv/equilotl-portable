"""Collect verbatim notices from the locked Go module cache, with no local paths."""
import hashlib
import json
import pathlib
import shutil
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[1]
output = ROOT / "third_party"
output.mkdir(exist_ok=True)
subprocess.run(["go", "mod", "download", "all"], cwd=ROOT, check=True)
raw = subprocess.check_output(["go", "list", "-m", "-json", "all"], cwd=ROOT, text=True)
decoder = json.JSONDecoder(); modules = []
while raw.strip():
    item, end = decoder.raw_decode(raw.lstrip()); raw = raw.lstrip()[end:]
    if not item.get("Main"):
        modules.append(item)
inventory = []
for module in sorted(modules, key=lambda m: m["Path"]):
    directory = pathlib.Path(module["Dir"])
    key = module["Path"].replace("/", "__") + "@" + module["Version"]
    notices = []
    for source in sorted(directory.rglob("*")):
        if not source.is_file() or not any(word in source.name.lower() for word in ("license", "licence", "copying", "notice", "copyright")):
            continue
        relative = source.relative_to(directory)
        dest = output / "licenses" / key / relative
        dest.parent.mkdir(parents=True, exist_ok=True); shutil.copyfile(source, dest)
        notices.append({"file": dest.relative_to(ROOT).as_posix(), "sha256": hashlib.sha256(dest.read_bytes()).hexdigest()})
    if not notices:
        raise SystemExit("No license notice found for " + module["Path"])
    inventory.append({"module": module["Path"], "version": module["Version"], "notices": notices})
# The Go runtime's BSD notice is needed for binary redistribution too.
goroot = pathlib.Path(subprocess.check_output(["go", "env", "GOROOT"], text=True).strip())
shutil.copyfile(goroot / "LICENSE", output / "Go-LICENSE.txt")
(output / "modules.json").write_text(json.dumps(inventory, indent=2) + "\n", encoding="utf-8", newline="\n")
print("Collected notices for", len(inventory), "locked modules and the Go runtime")
