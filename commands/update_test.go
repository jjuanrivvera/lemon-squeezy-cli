package commands

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/update"
	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/version"
)

// updateTestArchive builds the platform-correct release archive containing binName.
func updateTestArchive(t *testing.T, binName string) []byte {
	t.Helper()
	if runtime.GOOS == "windows" {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		w, err := zw.Create(binName)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte("NEW")); err != nil {
			t.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: binName, Mode: 0o755, Size: 3, Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("NEW")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestUpdateCommand drives `update` (against a stub GitHub) and `update check` through a
// temporary binary — never the test binary.
func TestUpdateCommand(t *testing.T) {
	binName := "lsqueezy"
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		binName += ".exe"
		ext = "zip"
	}
	archive := updateTestArchive(t, binName)
	asset := "lemon-squeezy-cli_0.9.9_" + runtime.GOOS + "_" + runtime.GOARCH + "." + ext
	sum := sha256.Sum256(archive)
	checksums := hex.EncodeToString(sum[:]) + "  " + asset + "\n"

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		host := "http://" + r.Host
		_ = json.NewEncoder(w).Encode(update.Release{TagName: "v0.9.9", Assets: []update.Asset{
			{Name: asset, BrowserDownloadURL: host + "/a"},
			{Name: "checksums.txt", BrowserDownloadURL: host + "/c"},
		}})
	})
	mux.HandleFunc("/a", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive) })
	mux.HandleFunc("/c", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(checksums)) })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	tmpExe := filepath.Join(t.TempDir(), "lsqueezy")
	if err := os.WriteFile(tmpExe, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}

	origFactory, origVer := newUpdater, version.Version
	newUpdater = func(v string) *update.Updater {
		u := update.NewUpdaterWithBaseURL(v, srv.URL)
		u.ExecutablePath = tmpExe
		return u
	}
	version.Version = "0.5.0"
	t.Cleanup(func() { newUpdater = origFactory; version.Version = origVer })

	up := newUpdateCmd()
	up.SetArgs([]string{})
	if err := up.Execute(); err != nil {
		t.Fatalf("update: %v", err)
	}
	if got, _ := os.ReadFile(tmpExe); string(got) != "NEW" {
		t.Errorf("binary not replaced: %q", got)
	}

	ck := newUpdateCmd()
	ck.SetArgs([]string{"check"})
	if err := ck.Execute(); err != nil {
		t.Fatalf("update check: %v", err)
	}
}

// TestUpdateCommand_AlreadyLatest covers the "no newer release" branches (current == latest).
func TestUpdateCommand_AlreadyLatest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(update.Release{TagName: "v0.9.9", Assets: []update.Asset{{Name: "x"}}})
	}))
	defer srv.Close()

	origFactory, origVer := newUpdater, version.Version
	newUpdater = func(v string) *update.Updater { return update.NewUpdaterWithBaseURL(v, srv.URL) }
	version.Version = "0.9.9" // equals the server's latest → "you are on the latest version"
	t.Cleanup(func() { newUpdater = origFactory; version.Version = origVer })

	up := newUpdateCmd()
	up.SetArgs([]string{})
	if err := up.Execute(); err != nil {
		t.Fatalf("update: %v", err)
	}

	ck := newUpdateCmd()
	ck.SetArgs([]string{"check"})
	if err := ck.Execute(); err != nil {
		t.Fatalf("update check: %v", err)
	}
}
