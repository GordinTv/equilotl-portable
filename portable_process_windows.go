package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type portableProcessState uint8

const (
	processInconclusive portableProcessState = iota
	processRunning
	processExited
)

type portableProcessObservation struct {
	PID        uint32
	Name       string
	Executable string
	State      portableProcessState
	QueryError error
}

type portableProcessBlock struct {
	Process portableProcessObservation
	Reason  string
}

type portableProcessReport struct {
	Blocking     []portableProcessBlock
	Inconclusive []portableProcessObservation
	ScanError    error
}

func portableVersionDir(di *DiscordInstall) string { return filepath.Dir(filepath.Dir(di.appPath)) }
func portableLauncherPath(di *DiscordInstall) string {
	return filepath.Join(filepath.Dir(filepath.Dir(portableVersionDir(di))), "discord-ptb-portable.exe")
}

// QueryFullProcessImageName normally returns a DOS path, but long/UNC paths
// can also carry an NT prefix. Require an absolute path and a directory boundary.
func normalizedProcessPath(p string) string {
	p = strings.ReplaceAll(p, "/", `\`)
	switch {
	case strings.HasPrefix(strings.ToLower(p), `\\?\unc\`):
		p = `\\` + p[8:]
	case strings.HasPrefix(p, `\\?\`):
		p = p[4:]
	case strings.HasPrefix(p, `\??\`):
		p = p[4:]
	}
	if !filepath.IsAbs(p) {
		return ""
	}
	return strings.ToLower(filepath.Clean(p))
}

func processWithin(p, root string) bool {
	p = normalizedProcessPath(p)
	root = normalizedProcessPath(root)
	return p != "" && root != "" && strings.HasPrefix(p, strings.TrimRight(root, `\`)+`\`)
}

func relevantProcessName(name string) bool {
	switch strings.ToLower(name) {
	case "discord.exe", "discordptb.exe", "discordcanary.exe", "discorddevelopment.exe", "discord-ptb-portable.exe", "electron.exe", "crashpad_handler.exe":
		return true
	}
	return false
}

func assessPortableProcesses(di *DiscordInstall, self uint32, observations []portableProcessObservation) portableProcessReport {
	var report portableProcessReport
	version, launcher := portableVersionDir(di), normalizedProcessPath(portableLauncherPath(di))
	for _, p := range observations {
		if p.PID == self || p.State == processExited {
			continue
		}
		inVersion := processWithin(p.Executable, version)
		isLauncher := launcher != "" && normalizedProcessPath(p.Executable) == launcher
		if p.State != processRunning || normalizedProcessPath(p.Executable) == "" {
			if relevantProcessName(p.Name) || inVersion || isLauncher {
				report.Inconclusive = append(report.Inconclusive, p)
			}
			continue
		}
		if inVersion {
			report.Blocking = append(report.Blocking, portableProcessBlock{p, "active process executing from the selected version; it may use its resources/app files"})
		} else if isLauncher {
			report.Blocking = append(report.Blocking, portableProcessBlock{p, "active launcher of this portable installation; it can start or update the selected host"})
		}
	}
	sort.Slice(report.Blocking, func(i, j int) bool { return report.Blocking[i].Process.PID < report.Blocking[j].Process.PID })
	return report
}

func (r portableProcessReport) warnings() string {
	var b strings.Builder
	if r.ScanError != nil {
		fmt.Fprintf(&b, "Process enumeration inconclusive: %v\n", r.ScanError)
	}
	for _, p := range r.Inconclusive {
		exe := p.Executable
		if exe == "" {
			exe = "<unavailable; not attributed to this installation>"
		}
		fmt.Fprintf(&b, "Inconclusive process: %s | PID: %d\nExecutable: %s\nReason: %v\n", p.Name, p.PID, exe, p.QueryError)
	}
	if b.Len() > 0 {
		b.WriteString("An inconclusive query is not proof that Discord Lite is running. Exclusive access to this installation's executable files is required before modifying files.\n")
	}
	return b.String()
}

func (r portableProcessReport) blockingError() error {
	if len(r.Blocking) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString("Discord Lite installation is in use (confirmed running processes). Close the listed processes before installing or uninstalling.\n")
	for _, block := range r.Blocking {
		p := block.Process
		fmt.Fprintf(&b, "\nProcess: %s | PID: %d\nExecutable: %s\nReason: %s\n", p.Name, p.PID, p.Executable, block.Reason)
	}
	if warning := r.warnings(); warning != "" {
		b.WriteString("\n" + warning)
	}
	return errors.New(b.String())
}

// Poll twice using the same process handle. In particular, failure to query
// an image after the process exits is not reported as access denial/running.
func observePortableProcess(p portableProcessObservation, poll func() (portableProcessState, error), image func() (string, error)) portableProcessObservation {
	p.State, p.QueryError = poll()
	if p.State != processRunning {
		return p
	}
	p.Executable, p.QueryError = image()
	state, err := poll()
	if state == processExited {
		p.State = processExited
		p.QueryError = nil
		return p
	}
	if p.QueryError != nil {
		p.State = processInconclusive
		return p
	}
	if state != processRunning {
		p.State = processInconclusive
		p.QueryError = err
		return p
	}
	if normalizedProcessPath(p.Executable) == "" {
		p.State = processInconclusive
		p.QueryError = errors.New("process did not return an absolute executable path")
	} else {
		p.Name = filepath.Base(p.Executable)
	}
	return p
}

func inspectWindowsProcess(pid uint32, name string) portableProcessObservation {
	p := portableProcessObservation{PID: pid, Name: name}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, pid)
	if err != nil {
		p.QueryError = err
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) || errors.Is(err, windows.ERROR_NOT_FOUND) {
			p.State = processExited
		}
		return p
	}
	defer windows.CloseHandle(h)
	return observePortableProcess(p, func() (portableProcessState, error) {
		wait, e := windows.WaitForSingleObject(h, 0)
		if e != nil {
			return processInconclusive, e
		}
		switch wait {
		case windows.WAIT_OBJECT_0:
			return processExited, nil
		case uint32(windows.WAIT_TIMEOUT):
			return processRunning, nil
		}
		return processInconclusive, fmt.Errorf("unexpected process wait result %d", wait)
	}, func() (string, error) {
		buf := make([]uint16, 32768)
		size := uint32(len(buf))
		e := windows.QueryFullProcessImageName(h, 0, &buf[0], &size)
		if e != nil {
			return "", e
		}
		return windows.UTF16ToString(buf[:size]), nil
	})
}

func collectPortableProcesses(di *DiscordInstall) portableProcessReport {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return portableProcessReport{ScanError: err}
	}
	defer windows.CloseHandle(snapshot)
	self := uint32(os.Getpid())
	var observations []portableProcessObservation
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	err = windows.Process32First(snapshot, &entry)
	for err == nil {
		// Inspect every executable, including renamed Electron helpers. Never infer
		// membership from the name, parent PID or a terminated launcher alone.
		if entry.ProcessID != 0 && entry.ProcessID != self {
			observations = append(observations, inspectWindowsProcess(entry.ProcessID, windows.UTF16ToString(entry.ExeFile[:])))
		}
		err = windows.Process32Next(snapshot, &entry)
	}
	report := assessPortableProcesses(di, self, observations)
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		report.ScanError = err
	}
	return refreshPortableBlockers(di, self, report, inspectWindowsProcess)
}

// Processes can also exit while the rest of the snapshot is being inspected.
// Recheck attributed blockers and their paths immediately before the decision;
// a reused PID must not inherit its predecessor's attribution.
func refreshPortableBlockers(di *DiscordInstall, self uint32, report portableProcessReport, inspect func(uint32, string) portableProcessObservation) portableProcessReport {
	observations := append([]portableProcessObservation(nil), report.Inconclusive...)
	for _, block := range report.Blocking {
		observations = append(observations, inspect(block.Process.PID, block.Process.Name))
	}
	current := assessPortableProcesses(di, self, observations)
	current.ScanError = report.ScanError
	return current
}

func portableProcessDiagnostic(di *DiscordInstall) (string, error) {
	report := collectPortableProcesses(di)
	if err := report.blockingError(); err != nil {
		return err.Error(), err
	}
	return "No active process was attributed to the selected portable version or its launcher.\n" + report.warnings(), nil
}

// A failed/denied process query must not turn into a name-based false positive.
// Instead require exclusive read/write ACCESS handles (without writing bytes)
// on every executable in this version
// and its shared launcher. This also catches inaccessible/renamed helpers and
// physical file aliases used from another path. Handles remain held until commit.
func leasePortableExecutable(di *DiscordInstall) (func(), error) {
	primary := filepath.Join(portableVersionDir(di), windowsNames[di.branch]+".exe")
	// Always probe the host even if it disappears while the directory walk runs.
	files := []string{primary}
	err := filepath.WalkDir(portableVersionDir(di), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Type()&os.ModeSymlink != 0 {
			if err := plainPath(p, true); err != nil {
				return err
			}
		}
		if !d.IsDir() {
			switch strings.ToLower(filepath.Ext(p)) {
			case ".exe", ".com", ".scr":
				if err := plainPath(p, false); err != nil {
					return err
				}
				if !strings.EqualFold(p, primary) {
					files = append(files, p)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot validate executable safety boundary: %w", err)
	}
	launcher := portableLauncherPath(di)
	if _, err = os.Lstat(launcher); err == nil {
		if err = plainPath(launcher, false); err != nil {
			return nil, err
		}
		files = append(files, launcher)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	selfPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	selfInfo, err := os.Stat(selfPath)
	if err != nil {
		return nil, err
	}
	var handles []windows.Handle
	release := func() {
		for i := len(handles) - 1; i >= 0; i-- {
			windows.CloseHandle(handles[i])
		}
	}
	for _, file := range files {
		info, e := os.Stat(file)
		if e != nil {
			release()
			return nil, e
		}
		if os.SameFile(info, selfInfo) {
			continue
		} // The installer itself is not Discord.
		p, e := windows.UTF16PtrFromString(file)
		if e != nil {
			release()
			return nil, e
		}
		// A mapped executable can still accept an exclusive READ handle after
		// its loader closes the original file handle. Requesting write access
		// also checks for an existing image section. Never write through h.
		h, e := windows.CreateFile(p, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
		if e != nil {
			release()
			return nil, fmt.Errorf("cannot safely modify this portable installation\nFile: %s\nReason: exclusive executable access probe unavailable (mapped image, file in use, read-only file or access denied); no process identity is inferred from this error: %w\n%s", file, e, di.portableProcessWarnings)
		}
		handles = append(handles, h)
	}
	return release, nil
}
