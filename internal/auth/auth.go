// Package auth stores the API key in the OS keyring, with an encrypted-file fallback for
// headless environments. Tokens never touch the config file or the repo.
package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

// service is the keyring service name; the profile is the keyring "user".
const service = "lemon-squeezy-cli"

// ErrNotFound indicates no token is stored for the profile.
var ErrNotFound = errors.New("no API key stored")

// Store is the token backend. The default uses the OS keyring; tests inject a fake.
type Store struct {
	// fallbackDir, when set, enables the encrypted-file fallback under this directory
	// (used when the keyring is unavailable, e.g. headless Linux/CI).
	fallbackDir string
	useFallback bool
}

// NewStore returns the default keyring-backed store. fallbackDir is where the encrypted
// file lives if the keyring errors.
func NewStore(fallbackDir string) *Store {
	return &Store{fallbackDir: fallbackDir}
}

// Set stores the key for a profile.
func (s *Store) Set(profile, key string) error {
	if err := keyring.Set(service, profile, key); err != nil {
		s.useFallback = true
		return s.setFallback(profile, key)
	}
	return nil
}

// Get retrieves the key for a profile, trying the keyring then the encrypted file.
func (s *Store) Get(profile string) (string, error) {
	key, err := keyring.Get(service, profile)
	if err == nil {
		return key, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		// Try the fallback before giving up — the key may have been written there.
		if fk, ferr := s.getFallback(profile); ferr == nil {
			return fk, nil
		}
		return "", ErrNotFound
	}
	// Keyring unavailable: fall back to the encrypted file.
	if fk, ferr := s.getFallback(profile); ferr == nil {
		return fk, nil
	}
	return "", ErrNotFound
}

// Delete removes the key for a profile from both backends.
func (s *Store) Delete(profile string) error {
	kerr := keyring.Delete(service, profile)
	ferr := s.deleteFallback(profile)
	if kerr != nil && !errors.Is(kerr, keyring.ErrNotFound) && ferr != nil {
		return kerr
	}
	return nil
}

// --- encrypted-file fallback ---
//
// The fallback derives an AES-256-GCM key from a machine-stable secret (the fallback dir
// path + a fixed salt). This is obfuscation-at-rest for headless boxes, not a substitute
// for a real secret service; the keyring remains the primary, preferred backend.

func (s *Store) fallbackPath(profile string) string {
	return filepath.Join(s.fallbackDir, "token-"+profile+".enc")
}

func (s *Store) deriveKey() []byte {
	h := sha256.Sum256([]byte("lemon-squeezy-cli-fallback-v1:" + s.fallbackDir))
	return h[:]
}

func (s *Store) setFallback(profile, key string) error {
	if s.fallbackDir == "" {
		return errors.New("keyring unavailable and no fallback directory configured")
	}
	if err := os.MkdirAll(s.fallbackDir, 0o700); err != nil {
		return err
	}
	block, err := aes.NewCipher(s.deriveKey())
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	ct := gcm.Seal(nonce, nonce, []byte(key), nil)
	enc := base64.StdEncoding.EncodeToString(ct)
	// #nosec G306 -- token file is intentionally 0600 (owner-only).
	return os.WriteFile(s.fallbackPath(profile), []byte(enc), 0o600)
}

func (s *Store) getFallback(profile string) (string, error) {
	if s.fallbackDir == "" {
		return "", ErrNotFound
	}
	raw, err := os.ReadFile(s.fallbackPath(profile)) // #nosec G304 -- path is derived, not user data
	if err != nil {
		return "", ErrNotFound
	}
	ct, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.deriveKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ct) < gcm.NonceSize() {
		return "", fmt.Errorf("corrupt token file")
	}
	nonce, body := ct[:gcm.NonceSize()], ct[gcm.NonceSize():]
	pt, err := gcm.Open(nil, nonce, body, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

func (s *Store) deleteFallback(profile string) error {
	if s.fallbackDir == "" {
		return nil
	}
	err := os.Remove(s.fallbackPath(profile))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
