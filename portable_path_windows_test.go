package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestWindowsShortPathIsNotALink(t *testing.T) {
	root := filepath.Join(t.TempDir(), "long directory for DOS alias")
	d := fixture(t, root, "1.0.1160")
	p, err := windows.UTF16PtrFromString(root)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]uint16, 32768)
	n, err := windows.GetShortPathName(p, &buf[0], uint32(len(buf)))
	if err != nil || n == 0 || n >= uint32(len(buf)) {
		t.Fatalf("GetShortPathName: %d %v", n, err)
	}
	short := windows.UTF16ToString(buf[:n])
	if err := plainPath(short, true); err != nil {
		t.Fatalf("ordinary DOS alias refused: %v", err)
	}
	found := ParseDiscord(short, "")
	if found == nil || !found.portable || found.version != d.version {
		t.Fatal("DOS alias installation not detected")
	}
	// A filesystem may have short-name creation disabled; still test its returned path.
	if strings.EqualFold(short, root) {
		t.Log("filesystem returned the long form; hosted TEMP also exercises existing aliases")
	}
}

func TestWindowsJunctionAncestorStillRefused(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	d := fixture(t, target, "1.0.1160")
	link := filepath.Join(root, "junction")
	cmd := exec.Command("cmd.exe", "/c", "mklink", "/J", link, target)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create fixture junction: %v: %s", err, output)
	}
	t.Cleanup(func() { _ = os.Remove(link) })
	if err := plainPath(link, true); err == nil {
		t.Fatal("junction accepted")
	}
	if err := plainPath(filepath.Join(link, filepath.Base(d.location()), "resources", "app"), true); err == nil {
		t.Fatal("junction ancestor accepted")
	}
	if ParseDiscord(link, "") != nil {
		t.Fatal("linked installation detected as safe")
	}
	if ParseDiscord(target, "") == nil {
		t.Fatal("ordinary target refused")
	}
}
