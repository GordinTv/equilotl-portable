package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func downloadFixture(t *testing.T) []byte {
	t.Helper()
	files := map[string]map[string]any{}
	payload := ""
	for _, name := range []string{"package.json", "patcher.js", "preload.js", "renderer.js"} {
		offset := len(payload)
		payload += "test"
		files[name] = map[string]any{"size": 4, "offset": strconv.Itoa(offset)}
	}
	header, _ := json.Marshal(map[string]any{"files": files})
	aligned := (len(header) + 3) &^ 3
	var b bytes.Buffer
	for _, n := range []uint32{4, uint32(aligned + 8), uint32(aligned + 4), uint32(len(header))} {
		if e := binary.Write(&b, binary.LittleEndian, n); e != nil {
			t.Fatal(e)
		}
	}
	b.Write(header)
	b.Write(make([]byte, aligned-len(header)))
	b.WriteString(payload)
	return b.Bytes()
}

func TestEquicordDownloadTransaction(t *testing.T) {
	oldDir, oldRelease, oldDev, oldHash := EquicordDirectory, ReleaseData, IsDevInstall, InstalledHash
	t.Cleanup(func() {
		EquicordDirectory = oldDir
		ReleaseData = oldRelease
		IsDevInstall = oldDev
		InstalledHash = oldHash
	})
	IsDevInstall = false
	valid := downloadFixture(t)
	for _, name := range []string{"success", "chunked", "truncated", "invalid-ASAR", "truncated-payload"} {
		t.Run(name, func(t *testing.T) {
			EquicordDirectory = filepath.Join(t.TempDir(), "equicord.asar")
			old := []byte("previous working build")
			if e := os.WriteFile(EquicordDirectory, old, 0644); e != nil {
				t.Fatal(e)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch name {
				case "truncated":
					w.Header().Set("Content-Length", "999999")
					w.Write(valid[:16])
				case "invalid-ASAR":
					w.Write([]byte("<html>upstream error</html>"))
				case "truncated-payload":
					w.Write(valid[:len(valid)-1])
				case "chunked":
					w.WriteHeader(200)
					w.(http.Flusher).Flush()
					w.Write(valid)
				default:
					w.Write(valid)
				}
			}))
			defer server.Close()
			release, _ := json.Marshal(map[string]any{"assets": []any{map[string]string{"name": "desktop.asar", "browser_download_url": server.URL}}})
			if e := json.Unmarshal(release, &ReleaseData); e != nil {
				t.Fatal(e)
			}
			err := installLatestBuilds()
			got, _ := os.ReadFile(EquicordDirectory)
			if name == "success" || name == "chunked" {
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, valid) {
					t.Fatal("download differs")
				}
				backup, _ := os.ReadFile(EquicordDirectory + ".previous")
				if !bytes.Equal(backup, old) {
					t.Fatal("previous build not preserved")
				}
			} else {
				if err == nil {
					t.Fatal("invalid download accepted")
				}
				if !bytes.Equal(got, old) {
					t.Fatal("failed download destroyed old build")
				}
			}
			matches, _ := filepath.Glob(filepath.Join(filepath.Dir(EquicordDirectory), ".equilotl-download-*"))
			if len(matches) != 0 {
				t.Fatal("temporary downloads leaked")
			}
		})
	}
}
