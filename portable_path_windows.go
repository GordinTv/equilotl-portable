package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Inspect every component without following reparse points. EvalSymlinks also
// expands ordinary DOS 8.3 names, so string inequality is not proof of a link.
func rejectPathLinks(abs string) error {
	for p := abs; ; p = filepath.Dir(p) {
		info, err := os.Lstat(p)
		if err != nil {
			return err
		}
		attrs, ok := info.Sys().(*syscall.Win32FileAttributeData)
		if !ok {
			return fmt.Errorf("cannot verify path attributes: %s", p)
		}
		if attrs.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
			return fmt.Errorf("linked path refused: %s", p)
		}
		if filepath.Dir(p) == p {
			return nil
		}
	}
}
