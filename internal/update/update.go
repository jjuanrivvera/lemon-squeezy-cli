// Package update self-updates the CLI binary from its GitHub releases. It is the
// fleet-standard self-updater: fetch the latest release, verify the archive against
// checksums.txt (mandatory), then atomically replace the running binary.
//
// PER-CLI: set githubRepo + binaryName below. Nothing else changes — findAssets is
// robust to the fleet's two archive-naming styles (with/without version, amd64/x86_64).
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
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	githubOwner = "jjuanrivvera"
	githubRepo  = "lemon-squeezy-cli" // e.g. "garminctl", "wootctl", "n8n-cli"
	binaryName  = "lsqueezy"          // the binary inside the archive, e.g. "n8nctl"
)

// Release / Asset mirror the subset of the GitHub releases API we use.
type Release struct {
	TagName    string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Result reports the outcome of an update attempt.
type Result struct {
	Updated     bool
	FromVersion string
	ToVersion   string
	Error       error
}

// Updater checks for and applies updates.
type Updater struct {
	CurrentVersion string
	HTTPClient     *http.Client
	ExecutablePath string // overridable for tests
	baseURL        string // GitHub API base; empty = the real API. Overridable in tests.
}

func NewUpdater(currentVersion string) *Updater {
	return &Updater{CurrentVersion: currentVersion, HTTPClient: &http.Client{Timeout: 30 * time.Second}}
}

// NewUpdaterWithBaseURL is NewUpdater with a custom GitHub API base URL. It exists so the
// `update` command is testable against an httptest server; pass "" for the real API.
func NewUpdaterWithBaseURL(currentVersion, baseURL string) *Updater {
	u := NewUpdater(currentVersion)
	u.baseURL = baseURL
	return u
}

// GetLatestRelease fetches the newest non-draft release.
func (u *Updater) GetLatestRelease(ctx context.Context) (*Release, error) {
	base := u.baseURL
	if base == "" {
		base = "https://api.github.com"
	}
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", base, githubOwner, githubRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", binaryName+"-updater")
	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// CheckAndUpdate updates to the latest release if it is newer than the running version.
func (u *Updater) CheckAndUpdate(ctx context.Context) *Result {
	res := &Result{FromVersion: u.CurrentVersion}
	if u.CurrentVersion == "dev" || u.CurrentVersion == "" {
		return res // never self-update a dev build
	}
	rel, err := u.GetLatestRelease(ctx)
	if err != nil {
		res.Error = fmt.Errorf("check for updates: %w", err)
		return res
	}
	if !isNewerVersion(rel.TagName, u.CurrentVersion) {
		return res // already current
	}
	res.ToVersion = strings.TrimPrefix(rel.TagName, "v")

	archive, checksums := u.findAssets(rel)
	if archive == nil {
		res.Error = fmt.Errorf("no compatible archive for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, rel.TagName)
		return res
	}
	// Checksum verification is mandatory. A release missing checksums.txt could mean
	// tampering or an incomplete publish; installing an unverified binary is unsafe.
	if checksums == nil {
		res.Error = fmt.Errorf("update aborted: release %s has no checksums.txt to verify integrity", rel.TagName)
		return res
	}
	blob, err := u.download(ctx, archive.BrowserDownloadURL)
	if err != nil {
		res.Error = fmt.Errorf("download archive: %w", err)
		return res
	}
	sums, err := u.download(ctx, checksums.BrowserDownloadURL)
	if err != nil {
		res.Error = fmt.Errorf("download checksums: %w", err)
		return res
	}
	if !verifyChecksum(blob, archive.Name, sums) {
		res.Error = fmt.Errorf("checksum verification failed for %s", archive.Name)
		return res
	}
	bin, err := extractBinary(blob, archive.Name)
	if err != nil {
		res.Error = fmt.Errorf("extract binary: %w", err)
		return res
	}
	if err := u.applyUpdate(bin); err != nil {
		res.Error = fmt.Errorf("apply update: %w", err)
		return res
	}
	res.Updated = true
	return res
}

// findAssets picks the archive for the current platform + the checksums file. It is
// deliberately tolerant of the fleet's two naming styles: with or without the version
// in the name, and amd64 vs x86_64 / arm64 vs aarch64. It matches by os token + arch
// token + archive extension, and skips the linux package formats (.deb/.rpm/.apk).
func (u *Updater) findAssets(rel *Release) (archive, checksums *Asset) {
	archTokens := []string{runtime.GOARCH}
	switch runtime.GOARCH {
	case "amd64":
		archTokens = []string{"amd64", "x86_64"}
	case "arm64":
		archTokens = []string{"arm64", "aarch64"}
	}
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	for i := range rel.Assets {
		a := &rel.Assets[i]
		if a.Name == "checksums.txt" {
			checksums = a
			continue
		}
		n := strings.ToLower(a.Name)
		if !strings.HasSuffix(n, ext) || !strings.Contains(n, runtime.GOOS) {
			continue
		}
		for _, t := range archTokens {
			if strings.Contains(n, t) {
				archive = a
				break
			}
		}
	}
	return archive, checksums
}

func (u *Updater) download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", binaryName+"-updater")
	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func verifyChecksum(data []byte, filename string, checksumsFile []byte) bool {
	want := ""
	for _, line := range strings.Split(string(checksumsFile), "\n") {
		f := strings.Fields(strings.TrimSpace(line))
		if len(f) >= 2 && f[1] == filename {
			want = f[0]
			break
		}
	}
	if want == "" {
		return false
	}
	sum := sha256.Sum256(data)
	return strings.EqualFold(want, hex.EncodeToString(sum[:]))
}

func extractBinary(archive []byte, archiveName string) ([]byte, error) {
	want := binaryName
	if runtime.GOOS == "windows" {
		want += ".exe"
	}
	if strings.HasSuffix(strings.ToLower(archiveName), ".zip") {
		zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) == want {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer func() { _ = rc.Close() }()
				return io.ReadAll(rc)
			}
		}
		return nil, fmt.Errorf("binary %s not found in archive", want)
	}
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag == tar.TypeReg && filepath.Base(h.Name) == want {
			return io.ReadAll(tr) //nolint:gosec // G110: release archives are small + checksum-verified upstream
		}
	}
	return nil, fmt.Errorf("binary %s not found in archive", want)
}

// applyUpdate atomically replaces the running binary, keeping a .bak to roll back on failure.
func (u *Updater) applyUpdate(newBinary []byte) error {
	execPath := u.ExecutablePath
	if execPath == "" {
		p, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolve executable: %w", err)
		}
		if p, err = filepath.EvalSymlinks(p); err != nil {
			return fmt.Errorf("resolve symlink: %w", err)
		}
		execPath = p
	}
	dir := filepath.Dir(execPath)
	tmp, err := os.CreateTemp(dir, "."+binaryName+"-update-*")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(newBinary); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write new binary: %w", err)
	}
	_ = tmp.Close()
	if err := os.Chmod(tmpPath, 0o755); err != nil { //nolint:gosec // G302: a CLI binary must be executable
		_ = os.Remove(tmpPath)
		return fmt.Errorf("chmod: %w", err)
	}
	backup := execPath + ".bak"
	if err := os.Rename(execPath, backup); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("backup current binary: %w", err)
	}
	if err := os.Rename(tmpPath, execPath); err != nil {
		_ = os.Rename(backup, execPath) // best-effort rollback
		_ = os.Remove(tmpPath)
		return fmt.Errorf("install new binary: %w", err)
	}
	_ = os.Remove(backup) // a stale .bak is harmless
	return nil
}

// isNewerVersion reports whether latest > current (semver, tolerant of a leading v and
// pre-release suffixes). A dev/empty current never triggers an update.
func isNewerVersion(latest, current string) bool {
	current = strings.TrimPrefix(current, "v")
	if current == "dev" || current == "" {
		return false
	}
	l, c := parseVersion(latest), parseVersion(current)
	for i := 0; i < 3; i++ {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

// IsNewer reports whether release tag `latest` is a newer version than `current`.
// Exported so an `update check` command can tell the user an update is available.
func IsNewer(latest, current string) bool { return isNewerVersion(latest, current) }

func parseVersion(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	var out [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		// A malformed component leaves out[i] == 0, which is the correct default.
		_, _ = fmt.Sscanf(strings.SplitN(parts[i], "-", 2)[0], "%d", &out[i])
	}
	return out
}
