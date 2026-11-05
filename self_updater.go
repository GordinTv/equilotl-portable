/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"errors"
	"fmt"
	"os"
)

// This independent fork must never replace itself with upstream Equilotl.
// Equicord build updates are separate and remain available through Repair.
var IsSelfOutdated = false
var SelfUpdateCheckDoneChan = make(chan bool, 1)

func init()                            { SelfUpdateCheckDoneChan <- false }
func GetInstallerDownloadLink() string { return "" }
func CanUpdateSelf() bool              { return false }
func UpdateSelf() error {
	return errors.New("automatic installer updates are disabled in this independent fork; download a verified release from this fork's repository")
}
func RelaunchSelf() error {
	attr := new(os.ProcAttr)
	attr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}

	var argv []string
	if len(os.Args) > 1 {
		argv = os.Args[1:]
	} else {
		argv = []string{}
	}

	Log.Debug("Restarting self with exe", os.Args[0], "and args", argv)

	proc, err := os.StartProcess(os.Args[0], argv, attr)
	if err != nil {
		return fmt.Errorf("Failed to start new process: %w", err)
	}

	if err = proc.Release(); err != nil {
		return fmt.Errorf("Failed to release new process: %w", err)
	}

	os.Exit(0)
	return nil
}
