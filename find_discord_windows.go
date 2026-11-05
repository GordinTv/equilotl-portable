/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	path "path/filepath"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var windowsNames = map[string]string{
	"stable": "Discord",
	"ptb":    "DiscordPTB",
	"canary": "DiscordCanary",
	"dev":    "DiscordDevelopment",
}

var killLock sync.Mutex

func ParseDiscordNew(p, branch string, isFlatpak bool) *DiscordInstall {
	return nil
}

func ParseDiscord(p, branch string) *DiscordInstall {
	p, err := path.Abs(p)
	if err != nil {
		return nil
	}
	// Accept the portable root, its app base, or an exact version directory.
	if plainPath(p, true) != nil {
		return nil
	}
	if plainPath(path.Join(p, "app"), true) == nil && !strings.HasPrefix(path.Base(p), "app-") {
		if d := ParseDiscord(path.Join(p, "app"), branch); d != nil && d.portable {
			return d
		}
	}
	if strings.HasPrefix(path.Base(p), "app-") {
		return parseWindowsVersion(path.Dir(p), path.Base(p), branch)
	}
	entries, err := os.ReadDir(p)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			Log.Warn("Error during readdir "+p+":", err)
		}
		return nil
	}

	var selected *DiscordInstall
	var latestVer []int
	for _, dir := range entries {
		if !dir.IsDir() || !strings.HasPrefix(dir.Name(), "app-") {
			continue
		}
		candidate := parseWindowsVersion(p, dir.Name(), branch)
		if candidate == nil {
			continue
		}
		ver := ParseAppVersion(dir.Name())
		if ver == nil {
			continue
		}
		if selected == nil || CompareAppVersion(ver, latestVer) > 0 {
			selected = candidate
			latestVer = ver
		}
	}

	return selected
}

func parseWindowsVersion(base, version, branch string) *DiscordInstall {
	if !strings.HasPrefix(version, "app-") || ParseAppVersion(version) == nil {
		return nil
	}
	resources := path.Join(base, version, "resources")
	if plainPath(resources, true) != nil {
		return nil
	}
	if branch == "" {
		branch = GetBranch(base)
		// A manually selected base may have an arbitrary name. Prefer host
		// metadata when its matching executable is actually present.
		var host struct{ ReleaseChannel string }
		if plainPath(path.Join(resources, "build_info.json"), false) == nil {
			if b, e := os.ReadFile(path.Join(resources, "build_info.json")); e == nil && json.Unmarshal(b, &host) == nil {
				if name, ok := windowsNames[host.ReleaseChannel]; ok && plainPath(path.Join(base, version, name+".exe"), false) == nil {
					branch = host.ReleaseChannel
				}
			}
		}
	}
	di := &DiscordInstall{path: base, branch: branch, appPath: path.Join(resources, "app"), version: strings.TrimPrefix(version, "app-")}
	if plainPath(path.Join(resources, "app.asar"), false) == nil || plainPath(path.Join(resources, "_app.asar"), false) == nil {
		di.isPatched = ExistsFile(path.Join(resources, "_app.asar"))
		return di
	}
	if !detectPortableApp(di.appPath) {
		return nil
	}
	if len(ParseAppVersion(version)) != 3 {
		return nil
	}
	var build struct {
		Version        string
		ReleaseChannel string
	}
	if plainPath(path.Join(resources, "build_info.json"), false) != nil {
		return nil
	}
	b, err := os.ReadFile(path.Join(resources, "build_info.json"))
	if err != nil || json.Unmarshal(b, &build) != nil || build.Version != di.version {
		return nil
	}
	exe, ok := windowsNames[build.ReleaseChannel]
	if !ok {
		return nil
	}
	if plainPath(path.Join(base, version, exe+".exe"), false) != nil {
		return nil
	}
	di.portable = true
	di.branch = build.ReleaseChannel
	di.isPatched = ExistsFile(path.Join(di.appPath, portableState))
	return di
}

func FindDiscords() []any {
	var discords []any

	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		Log.Error("%LOCALAPPDATA% is empty???????")
	}

	for branch, dirname := range windowsNames {
		if appData == "" {
			break
		}
		p := path.Join(appData, dirname)
		if discord := ParseDiscord(p, branch); discord != nil {
			Log.Debug("Found Discord install at ", p)
			discords = append(discords, discord)
		}
	}
	if roaming := os.Getenv("APPDATA"); roaming != "" {
		base := path.Join(roaming, "DC-Antilag", "app")
		entries, _ := os.ReadDir(base)
		for _, entry := range entries {
			if entry.IsDir() {
				if d := parseWindowsVersion(base, entry.Name(), ""); d != nil && d.portable {
					discords = append(discords, d)
				}
			}
		}
	}
	return discords
}

// Portable installs are never terminated by executable name: an official PTB
// may be open alongside them. Refuse while any process from this tree runs.
func portableProcessCheck(di *DiscordInstall) error {
	versionDir := path.Dir(path.Dir(di.appPath))
	current := parseWindowsVersion(path.Dir(versionDir), path.Base(versionDir), "")
	if current == nil || !current.portable || current.version != di.version || current.branch != di.branch {
		return errors.New("portable host changed since detection; refresh the installation list")
	}
	report := collectPortableProcesses(di)
	di.portableProcessWarnings = report.warnings()
	if di.portableProcessWarnings != "" {
		Log.Warn(di.portableProcessWarnings)
	}
	return report.blockingError()
}

func lockPortable(app string) (func(), error) {
	p, err := windows.UTF16PtrFromString(path.Join(app, ".equilotl-lock"))
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ|windows.GENERIC_WRITE|windows.DELETE, 0, nil, windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_DELETE_ON_CLOSE, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot lock portable install: %w", err)
	}
	return func() { windows.CloseHandle(h) }, nil
}

// ReplaceFile preserves the destination's DACL and creation metadata on NTFS.
// The staged file is already flushed and verified before this call.
func portableAtomicReplace(from, to string) error {
	if _, err := os.Lstat(to); errors.Is(err, os.ErrNotExist) {
		return os.Rename(from, to)
	} else if err != nil {
		return err
	}
	if err := plainPath(to, false); err != nil {
		return err
	}
	dst, err := windows.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	src, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	proc := windows.NewLazySystemDLL("kernel32.dll").NewProc("ReplaceFileW")
	r, _, e := proc.Call(uintptr(unsafe.Pointer(dst)), uintptr(unsafe.Pointer(src)), 0, 0, 0, 0)
	if r == 0 {
		return fmt.Errorf("ReplaceFile %s: %w", to, e)
	}
	return nil
}

func PreparePatch(di *DiscordInstall) {
	killLock.Lock()
	defer killLock.Unlock()

	name := windowsNames[di.branch]
	// Official PTB and Lite share the executable name. Only stop processes
	// whose image belongs to the selected installation.
	for _, pid := range findInstallProcessIDs(di) {
		proc, err := os.FindProcess(int(pid))
		if err != nil {
			Log.Warn("Failed to find process with pid", pid)
			continue
		}

		err = proc.Kill()
		if err != nil {
			Log.Warn("Failed to kill", name+":", err)
		} else {
			Log.Debug("Waiting for", name, "to exit")
			_, _ = proc.Wait()
		}
	}
}

func FixOwnership(_ string) error {
	return nil
}

// https://github.com/Vencord/Installer/issues/9

func CheckScuffedInstall() bool {
	username := os.Getenv("USERNAME")
	programData := os.Getenv("PROGRAMDATA")
	for _, discordName := range windowsNames {
		if ExistsFile(path.Join(programData, username, discordName)) {
			HandleScuffedInstall()
			return true
		}
	}
	return false
}

func findInstallProcessIDs(di *DiscordInstall) []uint32 {
	var ids []uint32
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(snapshot)

	procEntry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	err = windows.Process32First(snapshot, &procEntry)
	for {
		if err != nil {
			return ids
		}
		if strings.EqualFold(windows.UTF16ToString(procEntry.ExeFile[:]), windowsNames[di.branch]+".exe") {
			h, e := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, procEntry.ProcessID)
			if e == nil {
				buf := make([]uint16, 32768)
				size := uint32(len(buf))
				e = windows.QueryFullProcessImageName(h, 0, &buf[0], &size)
				windows.CloseHandle(h)
				if e == nil && strings.HasPrefix(strings.ToLower(windows.UTF16ToString(buf[:size])), strings.ToLower(path.Clean(di.path))+string(os.PathSeparator)) {
					ids = append(ids, procEntry.ProcessID)
				}
			}
		}
		err = windows.Process32Next(snapshot, &procEntry)
	}
}
