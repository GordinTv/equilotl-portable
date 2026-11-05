package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const portableState = ".equilotl-portable"
const portableLoader = ".equilotl-loader.cjs"

//go:embed portable_loader.cjs
var portableJS []byte

type portableManifest struct {
	Schema   int         `json:"schema"`
	Original string      `json:"originalSHA256"`
	Patched  string      `json:"patchedSHA256"`
	Loader   string      `json:"loaderSHA256"`
	Config   string      `json:"configSHA256"`
	Mode     os.FileMode `json:"mode"`
	Modified time.Time   `json:"modified"`
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// Fail closed on links/junctions, including ancestors, and on non-regular files.
func plainPath(p string, directory bool) error {
	abs, err := filepath.Abs(p)
	if err != nil {
		return err
	}
	if err := rejectPathLinks(abs); err != nil {
		return err
	}
	s, err := os.Lstat(p)
	if err != nil {
		return err
	}
	if directory && !s.IsDir() || !directory && !s.Mode().IsRegular() {
		return fmt.Errorf("unexpected file type: %s", p)
	}
	return nil
}

func portablePackage(app string) ([]byte, error) {
	if err := plainPath(app, true); err != nil {
		return nil, err
	}
	pkg := filepath.Join(app, "package.json")
	if err := plainPath(pkg, false); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	b, err := os.ReadFile(pkg)
	if errors.Is(err, os.ErrNotExist) {
		if _, e := readPortableState(app); e != nil {
			return nil, e
		}
		b, err = os.ReadFile(filepath.Join(app, portableState, "original.json"))
	}
	if err != nil {
		return nil, err
	}
	var p struct{ Name, Main, Type string }
	if json.Unmarshal(b, &p) != nil || p.Name != "discord" || p.Type == "module" {
		return nil, errors.New("invalid Discord CommonJS package")
	}
	if p.Main == portableLoader {
		if _, err := readPortableState(app); err != nil {
			return nil, err
		}
		b, err = os.ReadFile(filepath.Join(app, portableState, "original.json"))
		if err != nil {
			return nil, err
		}
		if json.Unmarshal(b, &p) != nil {
			return nil, errors.New("invalid original package")
		}
	}
	if filepath.ToSlash(p.Main) != "app_bootstrap/index.js" {
		return nil, errors.New("unsupported portable entry point")
	}
	for _, f := range []string{"app_bootstrap/index.js", "common/paths.js"} {
		if err := plainPath(filepath.Join(app, f), false); err != nil {
			return nil, err
		}
	}
	if err := plainPath(filepath.Join(app, "node_modules"), true); err != nil {
		return nil, err
	}
	return b, nil
}

func readPortableState(app string) (*portableManifest, error) {
	dir := filepath.Join(app, portableState)
	if err := plainPath(dir, true); err != nil {
		return nil, err
	}
	var m portableManifest
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	if len(entries) != 5 {
		return nil, errors.New("unexpected files in portable backup; refusing cleanup")
	}
	for _, f := range []string{"state.json", "original.json", "patched.json", "loader.cjs", "loader.json"} {
		if err := plainPath(filepath.Join(dir, f), false); err != nil {
			return nil, err
		}
	}
	b, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if err != nil {
		return nil, err
	}
	if json.Unmarshal(b, &m) != nil || m.Schema != 1 {
		return nil, errors.New("invalid portable backup manifest")
	}
	for f, h := range map[string]string{"original.json": m.Original, "patched.json": m.Patched, "loader.cjs": m.Loader, "loader.json": m.Config} {
		b, err = os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return nil, err
		}
		if digest(b) != h {
			return nil, fmt.Errorf("backup integrity failure: %s", f)
		}
	}
	return &m, nil
}

// Write, sync, close and verify before atomically replacing the destination.
func portableReplace(p string, b []byte, mode os.FileMode) error {
	f, err := os.CreateTemp(filepath.Dir(p), ".equilotl-write-")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Chmod(tmp, mode.Perm()); err != nil {
		return err
	}
	check, err := os.ReadFile(tmp)
	if err != nil {
		return err
	}
	if !bytes.Equal(check, b) {
		return errors.New("write verification failed")
	}
	return portableAtomicReplace(tmp, p)
}

func (di *DiscordInstall) portableStatus() string {
	if _, err := os.Stat(filepath.Join(di.appPath, portableState)); errors.Is(err, os.ErrNotExist) {
		return "Not installed"
	}
	m, err := readPortableState(di.appPath)
	if err != nil {
		return "Recovery required (backup invalid)"
	}
	b, _ := os.ReadFile(filepath.Join(di.appPath, "package.json"))
	l, _ := os.ReadFile(filepath.Join(di.appPath, portableLoader))
	if digest(b) == m.Patched && digest(l) == m.Loader {
		return "Equicord Installed"
	}
	return "Recovery required"
}

func (di *DiscordInstall) displayName() string {
	name := "Discord " + strings.Title(di.branch)
	if di.portable {
		name = "Discord Lite (Team AntiLag) / " + di.branch
	}
	if di.version != "" {
		name += " " + di.version
	}
	status := Ternary(di.isPatched, "Equicord Installed", "Not installed")
	if di.portable {
		status = di.portableStatus()
	}
	return name + " [" + status + "]"
}

func (di *DiscordInstall) location() string {
	if di.portable {
		return filepath.Dir(filepath.Dir(di.appPath))
	}
	return di.path
}

// Detection must still expose a damaged transaction so its recovery status is
// visible. Mutation always uses portablePackage and verifies the full backup.
func detectPortableApp(app string) bool {
	if _, err := portablePackage(app); err == nil {
		return true
	}
	if plainPath(app, true) != nil || plainPath(filepath.Join(app, "package.json"), false) != nil {
		return false
	}
	b, err := os.ReadFile(filepath.Join(app, "package.json"))
	if err != nil {
		return false
	}
	var p struct{ Name, Main, Type string }
	if json.Unmarshal(b, &p) != nil || p.Name != "discord" || p.Main != portableLoader || p.Type == "module" {
		return false
	}
	for _, f := range []string{"app_bootstrap/index.js", "common/paths.js"} {
		if plainPath(filepath.Join(app, f), false) != nil {
			return false
		}
	}
	return plainPath(filepath.Join(app, "node_modules"), true) == nil && plainPath(filepath.Join(app, portableState), true) == nil
}

func (di *DiscordInstall) portablePreflight() (func(), error) {
	if _, err := portablePackage(di.appPath); err != nil {
		return nil, err
	}
	release, err := lockPortable(di.appPath)
	if err != nil {
		return nil, err
	}
	if err = portableProcessCheck(di); err != nil {
		release()
		return nil, err
	}
	exeRelease, err := leasePortableExecutable(di)
	if err != nil {
		release()
		return nil, err
	}
	unlock := release
	release = func() { exeRelease(); unlock() }
	f, err := os.OpenFile(filepath.Join(di.appPath, "package.json"), os.O_RDWR, 0)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		release()
		return nil, err
	}
	if f != nil {
		f.Close()
	}
	return release, nil
}

func (di *DiscordInstall) patchPortable() (err error) {
	release, err := di.portablePreflight()
	if err != nil {
		return err
	}
	defer release()
	app := di.appPath
	state := filepath.Join(app, portableState)
	if _, err = os.Stat(EquicordDirectory); err != nil {
		return fmt.Errorf("Equicord unavailable: %w", err)
	}
	if _, e := os.Stat(state); e == nil {
		if err = di.restorePortable(); err != nil {
			return err
		}
	} else if !errors.Is(e, os.ErrNotExist) {
		return e
	}
	if _, e := os.Lstat(filepath.Join(app, portableLoader)); !errors.Is(e, os.ErrNotExist) {
		return errors.New("reserved portable loader path exists")
	}
	original, err := portablePackage(app)
	if err != nil {
		return err
	}
	if _, err = os.Stat(EquicordDirectory); err != nil {
		return fmt.Errorf("Equicord unavailable: %w", err)
	}
	var pkg map[string]json.RawMessage
	if err = json.Unmarshal(original, &pkg); err != nil {
		return err
	}
	pkg["main"], _ = json.Marshal(portableLoader)
	patched, _ := json.MarshalIndent(pkg, "", "  ")
	info, err := os.Stat(filepath.Join(app, "package.json"))
	if err != nil {
		return err
	}
	eq, err := filepath.Abs(EquicordDirectory)
	if err != nil {
		return err
	}
	config, _ := json.Marshal(map[string]string{"equicord": eq})
	m := portableManifest{Schema: 1, Original: digest(original), Patched: digest(patched), Loader: digest(portableJS), Config: digest(config), Mode: info.Mode(), Modified: info.ModTime()}
	manifest, _ := json.MarshalIndent(m, "", "  ")
	stage, err := os.MkdirTemp(app, ".equilotl-backup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	for f, b := range map[string][]byte{"original.json": original, "patched.json": patched, "loader.cjs": portableJS, "state.json": manifest, "loader.json": config} {
		if err = portableReplace(filepath.Join(stage, f), b, 0600); err != nil {
			return err
		}
	}
	if err = os.Rename(stage, state); err != nil {
		return err
	}
	if _, err = readPortableState(app); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if rollback := di.restorePortable(); rollback != nil {
				err = errors.Join(err, fmt.Errorf("rollback failed; verified backup retained at %s: %w", state, rollback))
			}
		}
	}()
	if err = di.portableCheckpoint("backup"); err != nil {
		return err
	}
	if err = portableReplace(filepath.Join(app, portableLoader), portableJS, 0644); err != nil {
		return err
	}
	if err = di.portableCheckpoint("loader"); err != nil {
		return err
	}
	current, e := os.ReadFile(filepath.Join(app, "package.json"))
	if e != nil {
		return e
	}
	if digest(current) != m.Original {
		return errors.New("package.json changed during patch; backup retained")
	}
	if err = portableReplace(filepath.Join(app, "package.json"), patched, info.Mode()); err != nil {
		return err
	}
	if err = di.portableCheckpoint("package"); err != nil {
		return err
	}
	di.isPatched = true
	return nil
}

// Also recovers interrupted install/uninstall. Unknown user changes are never overwritten.
func (di *DiscordInstall) restorePortable() error {
	if err := di.portableCheckpoint("restore"); err != nil {
		return err
	}
	app := di.appPath
	state := filepath.Join(app, portableState)
	m, err := readPortableState(app)
	if err != nil {
		return err
	}
	current, err := os.ReadFile(filepath.Join(app, "package.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if h := digest(current); err == nil && h != m.Original && h != m.Patched {
		return errors.New("package.json changed externally; backup retained, refusing overwrite")
	}
	loader := filepath.Join(app, portableLoader)
	if b, e := os.ReadFile(loader); e == nil {
		if plainPath(loader, false) != nil || digest(b) != m.Loader {
			return errors.New("loader changed externally; backup retained")
		}
	} else if !errors.Is(e, os.ErrNotExist) {
		return e
	}
	original, err := os.ReadFile(filepath.Join(state, "original.json"))
	if err != nil {
		return err
	}
	if err = portableReplace(filepath.Join(app, "package.json"), original, m.Mode); err != nil {
		return err
	}
	if err = os.Chtimes(filepath.Join(app, "package.json"), m.Modified, m.Modified); err != nil {
		return err
	}
	if err = os.Remove(loader); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// Remove only the exact verified files owned by this installer.
	tomb, err := os.MkdirTemp(app, ".equilotl-restored-")
	if err != nil {
		return err
	}
	os.Remove(tomb)
	if err = os.Rename(state, tomb); err != nil {
		return err
	}
	di.isPatched = false
	return os.RemoveAll(tomb)
}

func (di *DiscordInstall) portableCheckpoint(step string) error {
	if di.portableFault != nil {
		return di.portableFault(step)
	}
	return nil
}

func (di *DiscordInstall) unpatchPortable() error {
	release, err := di.portablePreflight()
	if err != nil {
		return err
	}
	defer release()
	if _, err = os.Stat(filepath.Join(di.appPath, portableState)); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return di.restorePortable()
}
