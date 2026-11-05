package main

import "testing"

func TestCommunityForkCannotInstallUpstreamSelfUpdate(t *testing.T) {
	old := IsSelfOutdated
	IsSelfOutdated = true
	defer func() { IsSelfOutdated = old }()
	if CanUpdateSelf() || GetInstallerDownloadLink() != "" {
		t.Fatal("community fork can download an upstream installer")
	}
	if UpdateSelf() == nil {
		t.Fatal("disabled self-update returned success")
	}
	select {
	case ready := <-SelfUpdateCheckDoneChan:
		SelfUpdateCheckDoneChan <- ready
		if ready {
			t.Fatal("disabled updater reported a successful remote check")
		}
	default:
		t.Fatal("disabled updater would block the CLI")
	}
}
