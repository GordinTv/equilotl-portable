package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestPortableProcessAttributionScenarios(t *testing.T) {
	di := &DiscordInstall{portable: true, branch: "ptb", version: "1.0.1160", appPath: `C:\Portable\DC-Antilag\app\app-1.0.1160\resources\app`}
	exe := filepath.Join(portableVersionDir(di), "DiscordPTB.exe")
	cases := []struct {
		name             string
		observations     []portableProcessObservation
		blocked, unknown int
	}{
		{"completely closed", nil, 0, 0},
		{"Lite running", []portableProcessObservation{{1, "DiscordPTB.exe", exe, processRunning, nil}}, 1, 0},
		{"portable launcher running", []portableProcessObservation{{2, "discord-ptb-portable.exe", portableLauncherPath(di), processRunning, nil}}, 1, 0},
		{"official Stable running", []portableProcessObservation{{3, "Discord.exe", `C:\Users\User\AppData\Local\Discord\app-1.0.1\Discord.exe`, processRunning, nil}}, 0, 0},
		{"official PTB running", []portableProcessObservation{{4, "DiscordPTB.exe", `C:\Users\User\AppData\Local\DiscordPTB\app-1.0.1160\DiscordPTB.exe`, processRunning, nil}}, 0, 0},
		{"renamed Electron helper inside target", []portableProcessObservation{{5, "utility-helper.exe", filepath.Join(portableVersionDir(di), "modules", "utility-helper.exe"), processRunning, nil}}, 1, 0},
		{"Electron helper outside target", []portableProcessObservation{{6, "electron.exe", `C:\Other\electron.exe`, processRunning, nil}}, 0, 0},
		{"same name different installation", []portableProcessObservation{{7, "DiscordPTB.exe", `D:\OtherLite\app\app-1.0.1160\DiscordPTB.exe`, processRunning, nil}}, 0, 0},
		{"sibling version", []portableProcessObservation{{8, "DiscordPTB.exe", `C:\Portable\DC-Antilag\app\app-1.0.1200\DiscordPTB.exe`, processRunning, nil}}, 0, 0},
		{"prefix collision", []portableProcessObservation{{9, "DiscordPTB.exe", `C:\Portable\DC-Antilag\app\app-1.0.1160-other\DiscordPTB.exe`, processRunning, nil}}, 0, 0},
		{"terminated during snapshot", []portableProcessObservation{{10, "DiscordPTB.exe", exe, processExited, nil}}, 0, 0},
		{"terminated launcher", []portableProcessObservation{{11, "discord-ptb-portable.exe", portableLauncherPath(di), processExited, nil}}, 0, 0},
		{"access denied is inconclusive", []portableProcessObservation{{12, "DiscordPTB.exe", "", processInconclusive, windows.ERROR_ACCESS_DENIED}}, 0, 1},
		{"mixed case", []portableProcessObservation{{13, "discordptb.EXE", strings.ToUpper(exe), processRunning, nil}}, 1, 0},
		{"installer itself", []portableProcessObservation{{777, "Equilotl.exe", filepath.Join(portableVersionDir(di), "Equilotl.exe"), processRunning, nil}}, 0, 0},
		{"long path prefix", []portableProcessObservation{{14, "DiscordPTB.exe", `\\?\` + exe, processRunning, nil}}, 1, 0},
		{"unrelated portable launcher", []portableProcessObservation{{15, "discord-ptb-portable.exe", `D:\OtherLite\discord-ptb-portable.exe`, processRunning, nil}}, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := assessPortableProcesses(di, 777, tc.observations)
			if len(r.Blocking) != tc.blocked || len(r.Inconclusive) != tc.unknown {
				t.Fatalf("got blockers=%d unknown=%d: %+v", len(r.Blocking), len(r.Inconclusive), r)
			}
			if (r.blockingError() != nil) != (tc.blocked > 0) {
				t.Fatal("wrong blocking decision")
			}
			if tc.blocked > 0 {
				message := r.blockingError().Error()
				for _, p := range r.Blocking {
					for _, text := range []string{p.Process.Name, fmt.Sprint(p.Process.PID), p.Process.Executable, p.Reason} {
						if !strings.Contains(message, text) {
							t.Fatalf("missing diagnostic %q", text)
						}
					}
				}
			}
			if tc.unknown > 0 && !strings.Contains(r.warnings(), "not proof") {
				t.Fatal("inconclusive result incorrectly attributed")
			}
		})
	}
}

func TestPortableProcessExitDuringImageQuery(t *testing.T) {
	polls := 0
	p := observePortableProcess(portableProcessObservation{PID: 55, Name: "DiscordPTB.exe"}, func() (portableProcessState, error) {
		polls++
		if polls == 1 {
			return processRunning, nil
		}
		return processExited, nil
	}, func() (string, error) { return "", windows.ERROR_ACCESS_DENIED })
	if p.State != processExited || p.QueryError != nil {
		t.Fatalf("exit race treated as active or permission failure: %+v", p)
	}
}

func TestPortableProcessQueryFailures(t *testing.T) {
	for _, name := range []string{"image denied", "wait denied", "invalid path"} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			p := observePortableProcess(portableProcessObservation{PID: 56, Name: "DiscordPTB.exe"}, func() (portableProcessState, error) {
				calls++
				if name == "wait denied" && calls == 2 {
					return processInconclusive, windows.ERROR_ACCESS_DENIED
				}
				return processRunning, nil
			}, func() (string, error) {
				if name == "image denied" {
					return "", windows.ERROR_ACCESS_DENIED
				}
				if name == "invalid path" {
					return "DiscordPTB.exe", nil
				}
				return `C:\Other\DiscordPTB.exe`, nil
			})
			if p.State != processInconclusive || p.QueryError == nil {
				t.Fatalf("query failure inferred a running target: %+v", p)
			}
		})
	}
}

func startFixtureProcess(t *testing.T, exe string) *exec.Cmd {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(os.Getenv("WINDIR"), "System32", "ping.exe"))
	if e != nil {
		t.Fatal(e)
	}
	put(t, exe, b)
	c := exec.Command(exe, "-n", "90", "127.0.0.1")
	if e = c.Start(); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { c.Process.Kill(); c.Wait() })
	return c
}

func TestPortableOpaqueActiveProcessLease(t *testing.T) {
	d := fixture(t, t.TempDir(), "1.0.1160")
	c := startFixtureProcess(t, filepath.Join(d.location(), "hidden-helper.exe"))
	actual := inspectWindowsProcess(uint32(c.Process.Pid), "hidden-helper.exe")
	if actual.State != processRunning {
		t.Fatal("helper not running", actual)
	}
	// Model a denied process-image query. The independent executable guard must
	// still reject the active target, even with no confirmed process attribution.
	if release, e := leasePortableExecutable(d); e == nil {
		release()
		t.Fatal("active mapped executable escaped file safety guard")
	}
}

func TestPortableLiveProcessScenarios(t *testing.T) {
	localEquicord(t)
	for _, name := range []string{"closed", "Lite", "launcher", "official Stable", "official PTB", "Electron helper", "same name outside", "sibling version"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			d := fixture(t, filepath.Join(root, "DC-Antilag", "app"), "1.0.1160")
			switch name {
			case "Lite":
				startFixtureProcess(t, filepath.Join(d.location(), "DiscordPTB.exe"))
			case "launcher":
				startFixtureProcess(t, portableLauncherPath(d))
			case "official Stable":
				startFixtureProcess(t, filepath.Join(root, "LocalAppData", "Discord", "app-1.0.1", "Discord.exe"))
			case "official PTB":
				startFixtureProcess(t, filepath.Join(root, "LocalAppData", "DiscordPTB", "app-1.0.1", "DiscordPTB.exe"))
			case "Electron helper":
				startFixtureProcess(t, filepath.Join(d.location(), "modules", "helper.exe"))
			case "same name outside":
				startFixtureProcess(t, filepath.Join(root, "OtherLite", "DiscordPTB.exe"))
			case "sibling version":
				startFixtureProcess(t, filepath.Join(d.path, "app-1.0.1200", "DiscordPTB.exe"))
			}
			before := tree(t, root)
			err := d.patchPortable()
			blocked := name == "Lite" || name == "launcher" || name == "Electron helper"
			if blocked {
				if err == nil {
					t.Fatal("active target allowed")
				}
				if !strings.Contains(err.Error(), "PID:") || !strings.Contains(err.Error(), "Executable:") || !strings.Contains(err.Error(), "Reason:") {
					t.Fatal("incomplete diagnostic", err)
				}
				assertTree(t, root, before)
			} else {
				if err != nil {
					t.Fatal("unrelated process blocked patch", err)
				}
				if err = d.unpatchPortable(); err != nil {
					t.Fatal(err)
				}
				assertTree(t, root, before)
			}
		})
	}
}

func TestPortableInconclusiveWithoutTargetActivity(t *testing.T) {
	d := fixture(t, t.TempDir(), "1.0.1160")
	r := assessPortableProcesses(d, 777, []portableProcessObservation{{PID: 123, Name: "DiscordPTB.exe", QueryError: windows.ERROR_ACCESS_DENIED}})
	if r.blockingError() != nil || len(r.Inconclusive) != 1 {
		t.Fatal("access error treated as proof of Lite membership")
	}
	d.portableProcessWarnings = r.warnings()
	release, e := leasePortableExecutable(d)
	if e != nil {
		t.Fatal("unrelated denied query must not block safe files", e)
	}
	release()
}

func TestPortableInconclusiveWithLockedTarget(t *testing.T) {
	d := fixture(t, t.TempDir(), "1.0.1160")
	p, _ := windows.UTF16PtrFromString(filepath.Join(d.location(), "DiscordPTB.exe"))
	h, e := windows.CreateFile(p, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer windows.CloseHandle(h)
	d.portableProcessWarnings = "Inconclusive process: DiscordPTB.exe | PID: 123; path unavailable"
	release, e := leasePortableExecutable(d)
	if e == nil {
		release()
		t.Fatal("inconclusive query bypassed resource lock")
	}
	if !strings.Contains(e.Error(), "PID: 123") || !strings.Contains(e.Error(), "File:") || !strings.Contains(e.Error(), "no process identity") {
		t.Fatal("resource failure wrongly reports a running process", e)
	}
}

func TestPortableTerminatedRealProcess(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "discord-ptb-portable.exe")
	b, e := os.ReadFile(filepath.Join(os.Getenv("WINDIR"), "System32", "ping.exe"))
	if e != nil {
		t.Fatal(e)
	}
	put(t, exe, b)
	c := exec.Command(exe, "-n", "1", "127.0.0.1")
	if e = c.Start(); e != nil {
		t.Fatal(e)
	}
	pid := uint32(c.Process.Pid)
	if e = c.Wait(); e != nil {
		t.Fatal(e)
	}
	p := inspectWindowsProcess(pid, "discord-ptb-portable.exe")
	if p.State != processExited {
		t.Fatalf("terminated launcher reported as running: %+v", p)
	}
}

func TestPortablePathNormalization(t *testing.T) {
	for _, p := range []string{`\\?\C:\Lite\DiscordPTB.exe`, `\??\C:\Lite\DiscordPTB.exe`, `c:/lite/DiscordPTB.exe`} {
		if normalizedProcessPath(p) != `c:\lite\discordptb.exe` {
			t.Fatal(p, normalizedProcessPath(p))
		}
	}
	if normalizedProcessPath(`\\?\UNC\server\share\DiscordPTB.exe`) != `\\server\share\discordptb.exe` {
		t.Fatal("UNC normalization failed")
	}
}

func TestPortableLeasePreservesExecutables(t *testing.T) {
	d := fixture(t, t.TempDir(), "1.0.1160")
	exe := filepath.Join(d.location(), "DiscordPTB.exe")
	before, e := os.Stat(exe)
	if e != nil {
		t.Fatal(e)
	}
	content, e := os.ReadFile(exe)
	if e != nil {
		t.Fatal(e)
	}
	release, e := leasePortableExecutable(d)
	if e != nil {
		t.Fatal(e)
	}
	release()
	after, e := os.Stat(exe)
	if e != nil {
		t.Fatal(e)
	}
	current, e := os.ReadFile(exe)
	if e != nil {
		t.Fatal(e)
	}
	if digest(content) != digest(current) || !before.ModTime().Equal(after.ModTime()) || before.Mode() != after.Mode() {
		t.Fatal("exclusive probe modified executable")
	}
}

func TestPortableReadOnlyExecutableIsNotReportedAsRunning(t *testing.T) {
	d := fixture(t, t.TempDir(), "1.0.1160")
	p, _ := windows.UTF16PtrFromString(filepath.Join(d.location(), "DiscordPTB.exe"))
	if e := windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_READONLY); e != nil {
		t.Fatal(e)
	}
	defer windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_NORMAL)
	release, e := leasePortableExecutable(d)
	if e == nil {
		release()
		t.Fatal("read-only executable bypassed required safety probe")
	}
	if strings.Contains(e.Error(), "confirmed running") || !strings.Contains(e.Error(), "read-only") {
		t.Fatal("permission uncertainty reported as a running process", e)
	}
}

func TestPortableSharedExecutableAlias(t *testing.T) {
	d := fixture(t, t.TempDir(), "1.0.1160")
	primary := filepath.Join(d.location(), "DiscordPTB.exe")
	b, e := os.ReadFile(filepath.Join(os.Getenv("WINDIR"), "System32", "ping.exe"))
	if e != nil {
		t.Fatal(e)
	}
	put(t, primary, b)
	alias := filepath.Join(t.TempDir(), "DiscordPTB.exe")
	if e = os.Link(primary, alias); e != nil {
		t.Fatal(e)
	}
	c := exec.Command(alias, "-n", "90", "127.0.0.1")
	if e = c.Start(); e != nil {
		t.Fatal(e)
	}
	defer func() { c.Process.Kill(); c.Wait() }()
	// Different path, but physically the same image: file-level safety still blocks.
	if release, e := leasePortableExecutable(d); e == nil {
		release()
		t.Fatal("shared image alias bypassed resource protection")
	}
}

func TestPortableHostDisappearedBeforeLease(t *testing.T) {
	d := fixture(t, t.TempDir(), "1.0.1160")
	if e := os.Remove(filepath.Join(d.location(), "DiscordPTB.exe")); e != nil {
		t.Fatal(e)
	}
	if release, e := leasePortableExecutable(d); e == nil {
		release()
		t.Fatal("missing host silently skipped")
	}
}

func TestPortableBlockerRevalidation(t *testing.T) {
	di := &DiscordInstall{appPath: `C:\Lite\app\app-1.0.1160\resources\app`}
	original := portableProcessObservation{PID: 123, Name: "DiscordPTB.exe", Executable: `C:\Lite\app\app-1.0.1160\DiscordPTB.exe`, State: processRunning}
	for _, name := range []string{"exited after enumeration", "PID reused outside installation", "query now denied"} {
		t.Run(name, func(t *testing.T) {
			report := assessPortableProcesses(di, 777, []portableProcessObservation{original})
			if len(report.Blocking) != 1 {
				t.Fatal("test fixture not attributed")
			}
			fresh := refreshPortableBlockers(di, 777, report, func(pid uint32, _ string) portableProcessObservation {
				if pid != 123 {
					t.Fatal("wrong process rechecked")
				}
				p := original
				switch name {
				case "exited after enumeration":
					p.State = processExited
				case "PID reused outside installation":
					p.Executable = `C:\Other\DiscordPTB.exe`
				default:
					p.State = processInconclusive
					p.Executable = ""
					p.QueryError = windows.ERROR_ACCESS_DENIED
				}
				return p
			})
			if fresh.blockingError() != nil {
				t.Fatal("stale attribution retained", fresh)
			}
			if name == "query now denied" && len(fresh.Inconclusive) != 1 {
				t.Fatal("denied recheck lost its diagnostic")
			}
		})
	}
}
