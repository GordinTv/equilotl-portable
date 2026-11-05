//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func portableAtomicReplace(from, to string) error { return os.Rename(from, to) }

func portableProcessCheck(_ *DiscordInstall) error {
	return errors.New("portable Discord is supported on Windows only")
}

func portableProcessDiagnostic(_ *DiscordInstall) (string, error) {
	return "", errors.New("portable process diagnostics require Windows")
}

func leasePortableExecutable(_ *DiscordInstall) (func(), error) {
	return nil, errors.New("portable Discord is supported on Windows only")
}
func lockPortable(_ string) (func(), error) {
	return nil, errors.New("portable Discord is supported on Windows only")
}

func rejectPathLinks(abs string) error {
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return err
	}
	if !strings.EqualFold(filepath.Clean(real), filepath.Clean(abs)) {
		return fmt.Errorf("linked path refused: %s", abs)
	}
	return nil
}
