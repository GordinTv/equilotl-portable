/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	path "path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type GithubRelease struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var ReleaseData GithubRelease
var GithubError error
var GithubDoneChan chan bool

var InstalledHash = "None"
var LatestHash = "Unknown"
var IsDevInstall bool

func GetGithubRelease(url, fallbackUrl string) (*GithubRelease, error) {
	Log.Debug("Fetching", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Log.Error("Failed to create Request", err)
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		Log.Error("Failed to send Request", err)
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode >= 300 {
		isRateLimitedOrBlocked := res.StatusCode == 401 || res.StatusCode == 403 || res.StatusCode == 429
		triedFallback := url == fallbackUrl

		if isRateLimitedOrBlocked && !triedFallback {
			Log.Error(fmt.Sprintf("Failed to fetch %s (status code %d). Trying fallback url %s", url, res.StatusCode, fallbackUrl))
			return GetGithubRelease(fallbackUrl, fallbackUrl)
		}

		err = errors.New(res.Status)
		Log.Error(url, "returned Non-OK status", GithubError)
		return nil, err
	}

	var data GithubRelease

	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		Log.Error("Failed to decode GitHub JSON Response", err)
		return nil, err
	}

	return &data, nil
}

func InitGithubDownloader() {
	GithubDoneChan = make(chan bool, 1)

	IsDevInstall = os.Getenv("EQUICORD_DEV_INSTALL") == "1"
	Log.Debug("Is Dev Install: ", IsDevInstall)
	if IsDevInstall {
		GithubDoneChan <- true
		return
	}

	go func() {
		// Make sure UI updates once the request either finished or failed
		defer func() {
			GithubDoneChan <- GithubError == nil
		}()

		data, err := GetGithubRelease(ReleaseUrl, ReleaseUrlFallback)
		if err != nil {
			GithubError = err
			return
		}

		ReleaseData = *data

		i := strings.LastIndex(data.Name, " ") + 1
		LatestHash = data.Name[i:]
		Log.Debug("Finished fetching GitHub Data")
		Log.Debug("Latest hash is", LatestHash, "Local Install is", Ternary(LatestHash == InstalledHash, "up to date!", "outdated!"))
	}()

	// either .asar file or directory with main.js file (in DEV)
	EquicordFile := EquicordDirectory

	stat, err := os.Stat(EquicordFile)
	if err != nil {
		return
	}

	// dev
	if stat.IsDir() {
		EquicordFile = path.Join(EquicordFile, "main.js")
	}

	// Check hash of installed version if exists
	b, err := os.ReadFile(EquicordFile)
	if err != nil {
		return
	}

	Log.Debug("Found existing Equicord Install. Checking for hash...")

	re := regexp.MustCompile(`// Equicord (\w+)`)
	match := re.FindSubmatch(b)
	if match != nil {
		InstalledHash = string(match[1])
		Log.Debug("Existing hash is", InstalledHash)

	} else {
		Log.Debug("Didn't find hash")

	}
}

func installLatestBuilds() (retErr error) {
	Log.Debug("Installing latest builds...")

	if IsDevInstall {
		Log.Debug("Skipping due to dev install")
		return
	}

	downloadUrl := ""
	for _, ass := range ReleaseData.Assets {
		if ass.Name == "desktop.asar" {
			downloadUrl = ass.DownloadURL
			break
		}
	}

	if downloadUrl == "" {
		retErr = errors.New("Didn't find desktop.asar download link")
		Log.Error(retErr)
		return
	}

	Log.Debug("Downloading desktop.asar")

	res, err := http.Get(downloadUrl)
	if res != nil {
		defer res.Body.Close()
	}
	if err == nil && res.StatusCode >= 300 {
		err = errors.New(res.Status)
	}
	if err != nil {
		Log.Error("Failed to download desktop.asar:", err)
		retErr = err
		return
	}
	out, err := os.CreateTemp(path.Dir(EquicordDirectory), ".equilotl-download-")
	if err != nil {
		Log.Error("Failed to create", EquicordDirectory+":", err)
		retErr = err
		return
	}
	defer os.Remove(out.Name())
	defer out.Close()
	read, err := io.Copy(out, res.Body)
	if err != nil {
		Log.Error("Failed to download to", EquicordDirectory+":", err)
		retErr = err
		return
	}
	if res.ContentLength >= 0 && read != res.ContentLength {
		err = fmt.Errorf("Unexpected end of input: expected %d bytes, received %d", res.ContentLength, read)
		Log.Error(err.Error())
		retErr = err
		return
	}
	if err = out.Sync(); err != nil {
		return err
	}
	if err = out.Close(); err != nil {
		return err
	}
	if err = validateEquicordDownload(out.Name()); err != nil {
		return err
	}
	// Preserve the previous build before committing a fully verified download.
	if previous, e := os.ReadFile(EquicordDirectory); e == nil {
		if err = portableReplace(EquicordDirectory+".previous", previous, 0644); err != nil {
			return err
		}
	} else if !errors.Is(e, os.ErrNotExist) {
		return e
	}
	if err = portableAtomicReplace(out.Name(), EquicordDirectory); err != nil {
		return err
	}

	_ = FixOwnership(EquicordDirectory)

	InstalledHash = LatestHash
	return
}

func validateEquicordDownload(p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	var header [16]byte
	if _, err = io.ReadFull(f, header[:]); err != nil {
		return err
	}
	size := binary.LittleEndian.Uint32(header[12:])
	if binary.LittleEndian.Uint32(header[:]) != 4 || size == 0 || size > 32*1024*1024 {
		return errors.New("invalid Equicord ASAR header")
	}
	b := make([]byte, size)
	if _, err = io.ReadFull(f, b); err != nil {
		return err
	}
	var index struct {
		Files map[string]struct {
			Size   int64
			Offset string
		}
	}
	if json.Unmarshal(b, &index) != nil {
		return errors.New("invalid Equicord ASAR index")
	}
	info, err := f.Stat()
	if err != nil {
		return err
	}
	payload := int64(8) + int64(binary.LittleEndian.Uint32(header[4:]))
	if payload < 16+int64(size) || payload > info.Size() {
		return errors.New("invalid ASAR payload boundary")
	}
	for _, name := range []string{"package.json", "patcher.js", "preload.js", "renderer.js"} {
		entry, ok := index.Files[name]
		offset, e := strconv.ParseInt(entry.Offset, 10, 64)
		if !ok || e != nil || offset < 0 || entry.Size <= 0 || offset > info.Size()-payload || entry.Size > info.Size()-payload-offset {
			return fmt.Errorf("Equicord ASAR missing or truncated %s", name)
		}
	}
	return nil
}
