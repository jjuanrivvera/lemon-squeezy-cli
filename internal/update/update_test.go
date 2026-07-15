package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIsNewerVersion(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v0.6.0", "0.5.3", true},
		{"0.6.0", "v0.6.0", false},
		{"v1.0.0", "v0.9.9", true},
		{"v0.2.0", "v0.10.0", false}, // 2 < 10, not lexical
		{"v0.10.0", "v0.2.0", true},
		{"v0.4.0", "v0.4.0", false},
		{"v0.4.1", "v0.4.0", true},
		{"v0.5.0-rc1", "0.4.0", true}, // pre-release suffix stripped
		{"v0.6.0", "dev", false},      // a dev build never self-updates
		{"v0.6.0", "", false},
	}
	for _, c := range cases {
		if got := IsNewer(c.latest, c.current); got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}

func TestFindAssets_BothFleetNamingStyles(t *testing.T) {
	goos, goarch := runtime.GOOS, runtime.GOARCH
	archAlt := goarch
	switch goarch {
	case "amd64":
		archAlt = "x86_64"
	case "arm64":
		archAlt = "aarch64"
	}
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	versioned := "repo_0.6.0_" + goos + "_" + goarch + "." + ext // most of the fleet
	noVersion := "repo_" + goos + "_" + archAlt + "." + ext      // canvas style (no version, x86_64)

	for _, name := range []string{versioned, noVersion} {
		rel := &Release{
			TagName: "v0.6.0",
			Assets: []Asset{
				{Name: "repo_0.6.0_linux_amd64.deb"}, // linux package formats must be skipped
				{Name: "repo_0.6.0_linux_amd64.rpm"},
				{Name: name, BrowserDownloadURL: "http://example/" + name},
				{Name: "checksums.txt", BrowserDownloadURL: "http://example/checksums.txt"},
			},
		}
		archive, checksums := NewUpdater("0.5.0").findAssets(rel)
		if archive == nil {
			t.Fatalf("no archive matched for %s/%s (asset %q)", goos, goarch, name)
		}
		if archive.Name != name {
			t.Errorf("picked %q, want %q", archive.Name, name)
		}
		if checksums == nil {
			t.Errorf("checksums.txt not found for asset set with %q", name)
		}
	}
}

func TestFindAssets_NoMatchReturnsNil(t *testing.T) {
	rel := &Release{Assets: []Asset{{Name: "repo_0.6.0_plan9_mips.tar.gz"}}}
	if a, _ := NewUpdater("0.5.0").findAssets(rel); a != nil {
		t.Errorf("expected no archive, got %q", a.Name)
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("release archive bytes")
	sum := sha256.Sum256(data)
	line := hex.EncodeToString(sum[:]) + "  repo_0.6.0_linux_amd64.tar.gz\n"
	checks := []byte("deadbeef  other.tar.gz\n" + line)

	if !verifyChecksum(data, "repo_0.6.0_linux_amd64.tar.gz", checks) {
		t.Error("valid checksum should verify")
	}
	if verifyChecksum([]byte("tampered"), "repo_0.6.0_linux_amd64.tar.gz", checks) {
		t.Error("tampered data must not verify")
	}
	if verifyChecksum(data, "not-listed.tar.gz", checks) {
		t.Error("a filename absent from checksums must not verify")
	}
}

func tarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExtractBinary_TarGz(t *testing.T) {
	name := binaryName
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	archive := tarGz(t, map[string]string{"README.md": "docs", name: "BINARY-CONTENT"})
	got, err := extractBinary(archive, "repo_0.6.0_linux_amd64.tar.gz")
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if string(got) != "BINARY-CONTENT" {
		t.Errorf("extracted %q, want BINARY-CONTENT", got)
	}
}

func TestExtractBinary_MissingBinary(t *testing.T) {
	archive := tarGz(t, map[string]string{"some-other-file": "abc"})
	if _, err := extractBinary(archive, "repo_0.6.0_linux_amd64.tar.gz"); err == nil {
		t.Error("expected an error when the binary is absent from the archive")
	}
}

func TestExtractBinary_Zip(t *testing.T) {
	name := binaryName
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	archive := zipArchive(t, name, "ZIP-BINARY")
	got, err := extractBinary(archive, "repo_0.9.9_windows_amd64.zip")
	if err != nil {
		t.Fatalf("extractBinary(zip): %v", err)
	}
	if string(got) != "ZIP-BINARY" {
		t.Errorf("extracted %q, want ZIP-BINARY", got)
	}
}

func TestApplyUpdate_ReplacesBinaryAndCleansBackup(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "binary")
	if err := os.WriteFile(exe, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	u := &Updater{CurrentVersion: "0.5.0", ExecutablePath: exe}
	if err := u.applyUpdate([]byte("NEW")); err != nil {
		t.Fatalf("applyUpdate: %v", err)
	}
	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "NEW" {
		t.Errorf("binary content = %q, want NEW", got)
	}
	if _, err := os.Stat(exe + ".bak"); !os.IsNotExist(err) {
		t.Error(".bak backup should be removed on success")
	}
}

func TestCheckAndUpdate_DevBuildShortCircuits(t *testing.T) {
	// A dev/empty version must never reach the network or replace the binary.
	for _, v := range []string{"dev", ""} {
		res := NewUpdater(v).CheckAndUpdate(context.Background())
		if res.Updated || res.Error != nil {
			t.Errorf("version %q: want no-op, got updated=%v err=%v", v, res.Updated, res.Error)
		}
	}
}

func zipArchive(t *testing.T, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestCheckAndUpdate_EndToEnd drives the full flow against a fake GitHub API + asset host:
// fetch latest → pick the platform asset → download → verify checksum → extract → replace.
func TestCheckAndUpdate_EndToEnd(t *testing.T) {
	binName := binaryName
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		binName += ".exe"
		ext = "zip"
	}
	var archive []byte
	if runtime.GOOS == "windows" {
		archive = zipArchive(t, binName, "NEW-BINARY")
	} else {
		archive = tarGz(t, map[string]string{binName: "NEW-BINARY"})
	}
	asset := "repo_0.9.9_" + runtime.GOOS + "_" + runtime.GOARCH + "." + ext
	sum := sha256.Sum256(archive)
	checksums := hex.EncodeToString(sum[:]) + "  " + asset + "\n"

	mux := http.NewServeMux()
	mux.HandleFunc("/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		host := "http://" + r.Host
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v0.9.9",
			Assets: []Asset{
				{Name: asset, BrowserDownloadURL: host + "/dl/archive"},
				{Name: "checksums.txt", BrowserDownloadURL: host + "/dl/checksums"},
			},
		})
	})
	mux.HandleFunc("/dl/archive", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive) })
	mux.HandleFunc("/dl/checksums", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(checksums)) })
	// The updater requests /repos/<owner>/<repo>/releases/latest; route any repos path to latest.
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			host := "http://" + r.Host
			_ = json.NewEncoder(w).Encode(Release{
				TagName: "v0.9.9",
				Assets: []Asset{
					{Name: asset, BrowserDownloadURL: host + "/dl/archive"},
					{Name: "checksums.txt", BrowserDownloadURL: host + "/dl/checksums"},
				},
			})
			return
		}
		http.NotFound(w, r)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	dir := t.TempDir()
	exe := filepath.Join(dir, "binary")
	if err := os.WriteFile(exe, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}

	u := &Updater{CurrentVersion: "0.5.0", HTTPClient: srv.Client(), ExecutablePath: exe, baseURL: srv.URL}
	res := u.CheckAndUpdate(context.Background())
	if res.Error != nil {
		t.Fatalf("CheckAndUpdate: %v", res.Error)
	}
	if !res.Updated {
		t.Fatal("expected Updated=true")
	}
	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "NEW-BINARY" {
		t.Errorf("binary content = %q, want NEW-BINARY", got)
	}
}

func TestCheckAndUpdate_AbortsWhenChecksumsMissing(t *testing.T) {
	asset := "repo_0.9.9_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	if runtime.GOOS == "windows" {
		asset = "repo_0.9.9_" + runtime.GOOS + "_" + runtime.GOARCH + ".zip"
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v0.9.9",
			Assets:  []Asset{{Name: asset, BrowserDownloadURL: "http://example/a"}}, // no checksums.txt
		})
	}))
	defer srv.Close()
	u := &Updater{CurrentVersion: "0.5.0", HTTPClient: srv.Client(), baseURL: srv.URL}
	if res := u.CheckAndUpdate(context.Background()); res.Error == nil {
		t.Fatal("expected an error when checksums.txt is absent from the release")
	}
}

func TestGetLatestRelease_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	u := &Updater{CurrentVersion: "0.5.0", HTTPClient: srv.Client(), baseURL: srv.URL}
	if _, err := u.GetLatestRelease(context.Background()); err == nil {
		t.Error("expected an error on HTTP 500")
	}
}

func TestGetLatestRelease_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("this is not json"))
	}))
	defer srv.Close()
	u := &Updater{CurrentVersion: "0.5.0", HTTPClient: srv.Client(), baseURL: srv.URL}
	if _, err := u.GetLatestRelease(context.Background()); err == nil {
		t.Error("expected a JSON decode error")
	}
}

func platformAsset() string {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return "repo_0.9.9_" + runtime.GOOS + "_" + runtime.GOARCH + "." + ext
}

func TestCheckAndUpdate_NoCompatibleAsset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v0.9.9",
			Assets:  []Asset{{Name: "repo_0.9.9_plan9_mips.tar.gz"}, {Name: "checksums.txt"}},
		})
	}))
	defer srv.Close()
	u := &Updater{CurrentVersion: "0.5.0", HTTPClient: srv.Client(), baseURL: srv.URL}
	if res := u.CheckAndUpdate(context.Background()); res.Error == nil {
		t.Fatal("expected an error when no asset matches the platform")
	}
}

func TestCheckAndUpdate_ChecksumMismatch(t *testing.T) {
	asset := platformAsset()
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		host := "http://" + r.Host
		_ = json.NewEncoder(w).Encode(Release{TagName: "v0.9.9", Assets: []Asset{
			{Name: asset, BrowserDownloadURL: host + "/a"},
			{Name: "checksums.txt", BrowserDownloadURL: host + "/c"},
		}})
	})
	mux.HandleFunc("/a", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("actual archive bytes")) })
	mux.HandleFunc("/c", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("deadbeef  " + asset + "\n")) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	u := &Updater{CurrentVersion: "0.5.0", HTTPClient: srv.Client(), baseURL: srv.URL}
	if res := u.CheckAndUpdate(context.Background()); res.Error == nil {
		t.Fatal("expected a checksum-mismatch error")
	}
}

func TestCheckAndUpdate_DownloadArchiveError(t *testing.T) {
	asset := platformAsset()
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		host := "http://" + r.Host
		_ = json.NewEncoder(w).Encode(Release{TagName: "v0.9.9", Assets: []Asset{
			{Name: asset, BrowserDownloadURL: host + "/missing"}, // 404s
			{Name: "checksums.txt", BrowserDownloadURL: host + "/c"},
		}})
	})
	mux.HandleFunc("/c", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("abc  " + asset + "\n")) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	u := &Updater{CurrentVersion: "0.5.0", HTTPClient: srv.Client(), baseURL: srv.URL}
	if res := u.CheckAndUpdate(context.Background()); res.Error == nil {
		t.Fatal("expected a download error when the archive 404s")
	}
}

func TestApplyUpdate_TempCreateError(t *testing.T) {
	// A target dir that does not exist makes the atomic temp-file create fail.
	u := &Updater{CurrentVersion: "0.5.0", ExecutablePath: filepath.Join(t.TempDir(), "missing-dir", "binary")}
	if err := u.applyUpdate([]byte("NEW")); err == nil {
		t.Error("expected an error when the target directory does not exist")
	}
}

func TestCheckAndUpdate_ExtractError(t *testing.T) {
	// The archive downloads and its checksum matches, but the binary is absent → extract fails.
	asset := platformAsset()
	var archive []byte
	if runtime.GOOS == "windows" {
		archive = zipArchive(t, "not-the-binary", "x")
	} else {
		archive = tarGz(t, map[string]string{"not-the-binary": "x"})
	}
	sum := sha256.Sum256(archive)
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		host := "http://" + r.Host
		_ = json.NewEncoder(w).Encode(Release{TagName: "v0.9.9", Assets: []Asset{
			{Name: asset, BrowserDownloadURL: host + "/a"},
			{Name: "checksums.txt", BrowserDownloadURL: host + "/c"},
		}})
	})
	mux.HandleFunc("/a", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive) })
	mux.HandleFunc("/c", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(hex.EncodeToString(sum[:]) + "  " + asset + "\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	u := &Updater{CurrentVersion: "0.5.0", HTTPClient: srv.Client(), baseURL: srv.URL, ExecutablePath: filepath.Join(t.TempDir(), "bin")}
	if res := u.CheckAndUpdate(context.Background()); res.Error == nil {
		t.Fatal("expected an extract error when the binary is absent from the archive")
	}
}
