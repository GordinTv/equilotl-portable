package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func put(t *testing.T, p string, b []byte) {
	t.Helper()
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p, b, 0644); e != nil {
		t.Fatal(e)
	}
}
func fixture(t *testing.T, base, version string) *DiscordInstall {
	t.Helper()
	root := filepath.Join(base, "app-"+version)
	app := filepath.Join(root, "resources", "app")
	put(t, filepath.Join(app, "package.json"), []byte("{\r\n  \"name\": \"discord\", \"main\": \"app_bootstrap/index.js\", \"description\": \"Discord Lite by Team AntiLag\"\r\n}\r\n"))
	put(t, filepath.Join(app, "app_bootstrap", "index.js"), []byte("global.originalLoads = (global.originalLoads || 0) + 1;"))
	put(t, filepath.Join(app, "common", "paths.js"), []byte("// original AntiLag paths"))
	if err := os.MkdirAll(filepath.Join(app, "node_modules"), 0755); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(root, "DiscordPTB.exe"), []byte("fixture executable"))
	build, _ := json.Marshal(map[string]string{"version": version, "releaseChannel": "ptb"})
	put(t, filepath.Join(root, "resources", "build_info.json"), build)
	di := ParseDiscord(base, "")
	if di == nil {
		t.Fatal("fixture not detected")
	}
	return di
}
func localEquicord(t *testing.T) string {
	t.Helper()
	old := EquicordDirectory
	t.Cleanup(func() { EquicordDirectory = old })
	eq := filepath.Join(t.TempDir(), "equicord.js")
	put(t, eq, []byte("// test Equicord"))
	EquicordDirectory = eq
	return eq
}
func tree(t *testing.T, root string) map[string]string {
	t.Helper()
	m := map[string]string{}
	e := filepath.WalkDir(root, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		r, _ := filepath.Rel(root, p)
		if d.IsDir() {
			m[r] = "directory"
		} else {
			b, e := os.ReadFile(p)
			if e != nil {
				return e
			}
			m[r] = digest(b)
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func assertTree(t *testing.T, root string, want map[string]string) {
	t.Helper()
	a, _ := json.Marshal(tree(t, root))
	b, _ := json.Marshal(want)
	if !bytes.Equal(a, b) {
		t.Fatalf("tree changed\ngot %s\nwant %s", a, b)
	}
}

func TestWindowsDetection(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "DC-Antilag", "app")
	d := fixture(t, base, "1.0.1160")
	before := tree(t, root)
	for _, p := range []string{base, filepath.Dir(base), filepath.Dir(filepath.Dir(d.appPath))} {
		if x := ParseDiscord(p, ""); x == nil || !x.portable || x.version != "1.0.1160" || x.branch != "ptb" {
			t.Fatalf("not detected: %s %#v", p, x)
		}
	}
	for _, p := range []string{root, filepath.Join(root, "missing"), d.appPath} {
		if ParseDiscord(p, "") != nil {
			t.Fatalf("accepted invalid path %s", p)
		}
	}
	assertTree(t, root, before)
	fixture(t, base, "1.0.1200")
	if x := ParseDiscord(base, ""); x.version != "1.0.1200" {
		t.Fatal(x.version)
	}
	put(t, filepath.Join(base, "app-1.0.1200", "resources", "build_info.json"), []byte(`{"version":"0.0.0","releaseChannel":"ptb"}`))
	if x := ParseDiscord(base, ""); x.version != "1.0.1160" {
		t.Fatal("invalid build selected")
	}
	t.Setenv("APPDATA", root)
	t.Setenv("LOCALAPPDATA", filepath.Join(root, "no-local"))
	found := false
	for _, x := range FindDiscords() {
		if x.(*DiscordInstall).portable {
			found = true
		}
	}
	if !found {
		t.Fatal("automatic portable discovery failed")
	}
}

func TestOfficialAsarLifecycle(t *testing.T) {
	localEquicord(t)
	for _, branch := range []string{"stable", "ptb", "canary", "dev"} {
		t.Run(branch, func(t *testing.T) {
			root := t.TempDir()
			res := filepath.Join(root, "app-1.0.1160", "resources")
			original := []byte("original official ASAR")
			put(t, filepath.Join(res, "app.asar"), original)
			di := ParseDiscord(root, branch)
			if di == nil || di.portable {
				t.Fatal("official detection regression")
			}
			if e := patchAppAsar(res, false); e != nil {
				t.Fatal(e)
			}
			if !ParseDiscord(root, branch).isPatched {
				t.Fatal("not patched")
			}
			if e := unpatchAppAsar(res, false); e != nil {
				t.Fatal(e)
			}
			b, _ := os.ReadFile(filepath.Join(res, "app.asar"))
			if !bytes.Equal(b, original) {
				t.Fatal("official original lost")
			}
		})
	}
}

func TestPortableLifecycleAndUpdate(t *testing.T) {
	localEquicord(t)
	root := t.TempDir()
	d := fixture(t, root, "1.0.1160")
	before := tree(t, root)
	pkg := filepath.Join(d.appPath, "package.json")
	info, _ := os.Stat(pkg)
	for i := 0; i < 3; i++ {
		if e := d.patchPortable(); e != nil {
			t.Fatal(e)
		}
		x := ParseDiscord(root, "")
		if x == nil || !x.isPatched || x.portableStatus() != "Equicord Installed" {
			t.Fatal("patch not detected")
		}
	}
	// A new Equicord location must update the config without losing the baseline.
	EquicordDirectory = filepath.Join(t.TempDir(), "new-equicord.js")
	put(t, EquicordDirectory, []byte("// updated Equicord"))
	if e := d.patchPortable(); e != nil {
		t.Fatal(e)
	}
	b, _ := os.ReadFile(filepath.Join(d.appPath, portableState, "loader.json"))
	if !bytes.Contains(b, []byte("new-equicord.js")) {
		t.Fatal("update did not update loader")
	}
	if e := d.unpatchPortable(); e != nil {
		t.Fatal(e)
	}
	assertTree(t, root, before)
	after, _ := os.Stat(pkg)
	if !after.ModTime().Equal(info.ModTime()) {
		t.Fatal("package timestamp not restored")
	}
	if e := d.unpatchPortable(); e != nil {
		t.Fatal(e)
	}
}

func TestPortableFailureAndRollback(t *testing.T) {
	localEquicord(t)
	for _, step := range []string{"backup", "loader", "package"} {
		t.Run(step, func(t *testing.T) {
			root := t.TempDir()
			d := fixture(t, root, "1.0.1160")
			before := tree(t, root)
			d.portableFault = func(s string) error {
				if s == step {
					return errors.New("injected patch failure")
				}
				return nil
			}
			if e := d.patchPortable(); e == nil {
				t.Fatal("failure not surfaced")
			}
			assertTree(t, root, before)
		})
	}
	t.Run("rollback failure and recovery", func(t *testing.T) {
		root := t.TempDir()
		d := fixture(t, root, "1.0.1160")
		before := tree(t, root)
		d.portableFault = func(s string) error {
			if s == "package" || s == "restore" {
				return errors.New("injected failure")
			}
			return nil
		}
		e := d.patchPortable()
		if e == nil || !strings.Contains(e.Error(), "rollback failed") {
			t.Fatal(e)
		}
		if _, e = readPortableState(d.appPath); e != nil {
			t.Fatal("backup not retained", e)
		}
		d.portableFault = nil
		if e = d.unpatchPortable(); e != nil {
			t.Fatal(e)
		}
		assertTree(t, root, before)
	})
}

func TestPortableInterruptedRecovery(t *testing.T) {
	localEquicord(t)
	for _, phase := range []string{"backup", "loader", "package"} {
		t.Run(phase, func(t *testing.T) {
			root := t.TempDir()
			d := fixture(t, root, "1.0.1160")
			before := tree(t, root)
			if e := d.patchPortable(); e != nil {
				t.Fatal(e)
			}
			if phase != "package" {
				b, _ := os.ReadFile(filepath.Join(d.appPath, portableState, "original.json"))
				put(t, filepath.Join(d.appPath, "package.json"), b)
			}
			if phase == "backup" {
				if e := os.Remove(filepath.Join(d.appPath, portableLoader)); e != nil {
					t.Fatal(e)
				}
			}
			d = ParseDiscord(root, "")
			if d == nil {
				t.Fatal("interrupted install not detected")
			}
			if e := d.unpatchPortable(); e != nil {
				t.Fatal(e)
			}
			assertTree(t, root, before)
		})
	}
}

func TestPortableCorruptionAndForeignChanges(t *testing.T) {
	localEquicord(t)
	for _, name := range []string{"backup", "package", "loader"} {
		t.Run(name, func(t *testing.T) {
			d := fixture(t, t.TempDir(), "1.0.1160")
			if e := d.patchPortable(); e != nil {
				t.Fatal(e)
			}
			p := filepath.Join(d.appPath, portableState, "original.json")
			if name == "package" {
				p = filepath.Join(d.appPath, "package.json")
			}
			if name == "loader" {
				p = filepath.Join(d.appPath, portableLoader)
			}
			put(t, p, []byte("external modification"))
			before := tree(t, d.appPath)
			if e := d.unpatchPortable(); e == nil {
				t.Fatal("corruption accepted")
			}
			assertTree(t, d.appPath, before)
		})
	}
}

func TestPortablePermissionDeniedAndExclusiveLock(t *testing.T) {
	localEquicord(t)
	d := fixture(t, t.TempDir(), "1.0.1160")
	before := tree(t, d.path)
	p, _ := windows.UTF16PtrFromString(filepath.Join(d.appPath, "package.json"))
	h, e := windows.CreateFile(p, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	if e = d.patchPortable(); e == nil {
		t.Fatal("accepted read-only shared file")
	}
	windows.CloseHandle(h)
	assertTree(t, d.path, before)
	release, e := lockPortable(d.appPath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = lockPortable(d.appPath); e == nil {
		t.Fatal("second concurrent writer allowed")
	}
	release()
	assertTree(t, d.path, before)
	// Windows read-only attribute rejects the O_RDWR preflight.
	if e = windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_READONLY); e != nil {
		t.Fatal(e)
	}
	defer windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_NORMAL)
	if e = d.patchPortable(); e == nil {
		t.Fatal("read-only package accepted")
	}
}

func TestPortableRunningProcess(t *testing.T) {
	localEquicord(t)
	d := fixture(t, t.TempDir(), "1.0.1160")
	exe := filepath.Join(d.path, "app-1.0.1160", "DiscordPTB.exe")
	// Harmless system sleep process renamed in a temporary fixture, never real Discord.
	b, e := os.ReadFile(filepath.Join(os.Getenv("WINDIR"), "System32", "ping.exe"))
	if e != nil {
		t.Fatal(e)
	}
	put(t, exe, b)
	c := exec.Command(exe, "-n", "15", "127.0.0.1")
	if e = c.Start(); e != nil {
		t.Fatal(e)
	}
	defer func() { c.Process.Kill(); c.Wait() }()
	time.Sleep(100 * time.Millisecond)
	before := tree(t, d.path)
	if e = d.patchPortable(); e == nil || !strings.Contains(e.Error(), "running") {
		t.Fatal(e)
	}
	assertTree(t, d.path, before)
}

func TestPortableVersionUpgrade(t *testing.T) {
	localEquicord(t)
	root := t.TempDir()
	old := fixture(t, root, "1.0.1160")
	if e := old.patchPortable(); e != nil {
		t.Fatal(e)
	}
	newest := fixture(t, root, "1.0.1200")
	if newest.isPatched || newest.version != "1.0.1200" {
		t.Fatal("new version state inherited")
	}
	if e := newest.patchPortable(); e != nil {
		t.Fatal(e)
	}
	if e := newest.unpatchPortable(); e != nil {
		t.Fatal(e)
	}
	if old.portableStatus() != "Equicord Installed" {
		t.Fatal("new version changed old backup")
	}
	if e := old.unpatchPortable(); e != nil {
		t.Fatal(e)
	}
}

func TestPortableNodeResolution(t *testing.T) {
	if _, e := exec.LookPath("node"); e != nil {
		t.Skip("node missing")
	}
	eq := localEquicord(t)
	d := fixture(t, t.TempDir(), "1.0.1160")
	// Exercise the exact public Equicord loader contract, in a fresh Node process.
	put(t, eq, []byte(`const path=require('path');const app=require('electron').app;const a=path.join(path.dirname(require.main.filename),'..','_app.asar');const pkg=require(path.join(a,'package.json'));require.main.filename=path.join(a,pkg.main);app.setAppPath(a);require(require.main.filename);global.equicordLoads=(global.equicordLoads||0)+1;`))
	if e := d.patchPortable(); e != nil {
		t.Fatal(e)
	}
	harness := filepath.Join(t.TempDir(), "harness.cjs")
	js := `const Module=require('module'),assert=require('assert'),path=require('path');const root=process.argv[2];let actual;const electron={app:{setAppPath:p=>actual=p}};const load=Module._load;Module._load=function(id,...args){if(id==='electron')return electron;return load.call(this,id,...args)};require.main.filename=path.join(root,'.equilotl-loader.cjs');require(path.join(root,'.equilotl-loader.cjs'));assert.equal(global.originalLoads,1);assert.equal(global.equicordLoads,1);assert.equal(actual,root);assert.equal(require.main.filename,path.join(root,'app_bootstrap/index.js'));require('fs');require('assert');console.log('PASS original=1 equicord=1 appPath=original main=original');`
	put(t, harness, []byte(js))
	out, e := exec.Command("node", harness, d.appPath).CombinedOutput()
	if e != nil {
		t.Fatalf("%v %s", e, out)
	}
	t.Log(string(out))
}

func TestPortableRecoveryMissingPackage(t *testing.T) {
	localEquicord(t)
	root := t.TempDir()
	d := fixture(t, root, "1.0.1160")
	before := tree(t, root)
	if e := d.patchPortable(); e != nil {
		t.Fatal(e)
	}
	if e := os.Remove(filepath.Join(d.appPath, "package.json")); e != nil {
		t.Fatal(e)
	}
	detected := ParseDiscord(root, "")
	if detected == nil || !detected.portable {
		t.Fatal("lost package transaction not detected")
	}
	if e := detected.unpatchPortable(); e != nil {
		t.Fatal(e)
	}
	assertTree(t, root, before)
}

func TestPortableUnknownBackupFilesAndDetection(t *testing.T) {
	localEquicord(t)
	d := fixture(t, t.TempDir(), "1.0.1160")
	if e := d.patchPortable(); e != nil {
		t.Fatal(e)
	}
	put(t, filepath.Join(d.appPath, portableState, "user-document.txt"), []byte("must not delete"))
	before := tree(t, d.path)
	detected := ParseDiscord(d.path, "")
	if detected == nil || !strings.Contains(detected.portableStatus(), "backup invalid") {
		t.Fatal("damaged backup not exposed")
	}
	if e := d.unpatchPortable(); e == nil {
		t.Fatal("unknown backup content deleted")
	}
	assertTree(t, d.path, before)
}

func TestPortableMissingEquicordDoesNotRemoveExistingPatch(t *testing.T) {
	localEquicord(t)
	d := fixture(t, t.TempDir(), "1.0.1160")
	if e := d.patchPortable(); e != nil {
		t.Fatal(e)
	}
	before := tree(t, d.path)
	EquicordDirectory = filepath.Join(t.TempDir(), "missing.asar")
	if e := d.patchPortable(); e == nil {
		t.Fatal("missing Equicord accepted")
	}
	assertTree(t, d.path, before)
}

func TestPortableExecutableLease(t *testing.T) {
	localEquicord(t)
	d := fixture(t, t.TempDir(), "1.0.1160")
	release, e := d.portablePreflight()
	if e != nil {
		t.Fatal(e)
	}
	p, _ := windows.UTF16PtrFromString(filepath.Join(d.location(), "DiscordPTB.exe"))
	h, e := windows.CreateFile(p, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, 0, 0)
	if e == nil {
		windows.CloseHandle(h)
		release()
		t.Fatal("executable not held exclusively")
	}
	release()
	h, e = windows.CreateFile(p, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	windows.CloseHandle(h)
}

func TestPortableACLPermissionsAndPreservation(t *testing.T) {
	localEquicord(t)
	d := fixture(t, t.TempDir(), "1.0.1160")
	pkg := filepath.Join(d.appPath, "package.json")
	before := tree(t, d.path)
	old, e := windows.GetNamedSecurityInfo(pkg, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if e != nil {
		t.Fatal(e)
	}
	oldACL, _, e := old.DACL()
	if e != nil {
		t.Fatal(e)
	}
	restore := func() {
		if e := windows.SetNamedSecurityInfo(pkg, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.UNPROTECTED_DACL_SECURITY_INFORMATION, nil, nil, oldACL, nil); e != nil {
			t.Error(e)
		}
	}
	defer restore()
	user, e := windows.GetCurrentProcessToken().GetTokenUser()
	if e != nil {
		t.Fatal(e)
	}
	sid := user.User.Sid.String()
	denied, e := windows.SecurityDescriptorFromString("D:P(D;;0x2;;;" + sid + ")(A;;FA;;;" + sid + ")")
	if e != nil {
		t.Fatal(e)
	}
	acl, _, e := denied.DACL()
	if e != nil {
		t.Fatal(e)
	}
	if e = windows.SetNamedSecurityInfo(pkg, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); e != nil {
		t.Fatal(e)
	}
	if e = d.patchPortable(); e == nil || !errors.Is(e, os.ErrPermission) {
		t.Fatalf("write-denying ACL not respected: %v", e)
	}
	restore()
	assertTree(t, d.path, before)
	sdBefore, e := windows.GetNamedSecurityInfo(pkg, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if e != nil {
		t.Fatal(e)
	}
	if e = d.patchPortable(); e != nil {
		t.Fatal(e)
	}
	if e = d.unpatchPortable(); e != nil {
		t.Fatal(e)
	}
	sdAfter, e := windows.GetNamedSecurityInfo(pkg, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if e != nil {
		t.Fatal(e)
	}
	if sdBefore.String() != sdAfter.String() {
		t.Fatalf("ACL changed: %s -> %s", sdBefore.String(), sdAfter.String())
	}
}

func TestOfficialStaleOriginalPreserved(t *testing.T) {
	root := t.TempDir()
	original := []byte("old original Discord ASAR")
	newHost := []byte("new original Discord ASAR")
	put(t, filepath.Join(root, "_app.asar"), original)
	put(t, filepath.Join(root, "app.asar"), newHost)
	cleaned, e := cleanupDesyncedPatchedInstall(root, false)
	if e != nil || !cleaned {
		t.Fatal(cleaned, e)
	}
	matches, _ := filepath.Glob(filepath.Join(root, ".equilotl-stale-*", "_app.asar"))
	if len(matches) != 1 {
		t.Fatal("original archive missing")
	}
	b, _ := os.ReadFile(matches[0])
	if !bytes.Equal(b, original) {
		t.Fatal("original backup differs")
	}
	b, _ = os.ReadFile(filepath.Join(root, "app.asar"))
	if !bytes.Equal(b, newHost) {
		t.Fatal("new host changed")
	}
}

func TestPortableInvalidLayoutAndStaleSelection(t *testing.T) {
	localEquicord(t)
	for _, name := range []string{"package-name", "main-traversal", "bootstrap-missing", "exe-missing", "channel-invalid", "build-version-changed"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			d := fixture(t, root, "1.0.1160")
			res := filepath.Dir(d.appPath)
			switch name {
			case "package-name":
				put(t, filepath.Join(d.appPath, "package.json"), []byte(`{"name":"unrelated","main":"app_bootstrap/index.js"}`))
			case "main-traversal":
				put(t, filepath.Join(d.appPath, "package.json"), []byte(`{"name":"discord","main":"../index.js"}`))
			case "bootstrap-missing":
				if e := os.Remove(filepath.Join(d.appPath, "app_bootstrap", "index.js")); e != nil {
					t.Fatal(e)
				}
			case "exe-missing":
				if e := os.Remove(filepath.Join(d.location(), "DiscordPTB.exe")); e != nil {
					t.Fatal(e)
				}
			case "channel-invalid":
				put(t, filepath.Join(res, "build_info.json"), []byte(`{"version":"1.0.1160","releaseChannel":"unrelated"}`))
			case "build-version-changed":
				put(t, filepath.Join(res, "build_info.json"), []byte(`{"version":"1.0.1200","releaseChannel":"ptb"}`))
			}
			before := tree(t, root)
			if ParseDiscord(root, "") != nil {
				t.Fatal("invalid layout accepted")
			}
			if e := d.patchPortable(); e == nil {
				t.Fatal("stale GUI/CLI selection accepted")
			}
			assertTree(t, root, before)
		})
	}
}
