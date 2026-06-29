package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestKeyringRoundTrip(t *testing.T) {
	keyring.MockInit() // in-memory keyring; no real OS Keychain
	s := NewStore(t.TempDir())

	_, err := s.Get("default")
	assert.ErrorIs(t, err, ErrNotFound)

	require.NoError(t, s.Set("default", "secret-key"))
	got, err := s.Get("default")
	require.NoError(t, err)
	assert.Equal(t, "secret-key", got)

	require.NoError(t, s.Delete("default"))
	_, err = s.Get("default")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestKeyringProfilesAreIsolated(t *testing.T) {
	keyring.MockInit()
	s := NewStore(t.TempDir())
	require.NoError(t, s.Set("a", "key-a"))
	require.NoError(t, s.Set("b", "key-b"))
	a, _ := s.Get("a")
	b, _ := s.Get("b")
	assert.Equal(t, "key-a", a)
	assert.Equal(t, "key-b", b)
}

func TestEncryptedFileFallbackRoundTrip(t *testing.T) {
	// Drive the fallback directly (keyring unavailable path).
	dir := t.TempDir()
	s := NewStore(dir)
	require.NoError(t, s.setFallback("default", "fallback-secret"))
	got, err := s.getFallback("default")
	require.NoError(t, err)
	assert.Equal(t, "fallback-secret", got)

	require.NoError(t, s.deleteFallback("default"))
	_, err = s.getFallback("default")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFallbackEncryptedAtRest(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	require.NoError(t, s.setFallback("default", "plaintext-token"))
	// The encrypted file must not contain the cleartext token.
	raw, err := os.ReadFile(s.fallbackPath("default"))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "plaintext-token")
}

func TestFallbackCorruptFile(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(s.fallbackPath("default"), []byte("!!!notbase64!!!"), 0o600))
	_, err := s.getFallback("default")
	assert.Error(t, err)
}

func TestGet_KeyringMissingFallsBackToFile(t *testing.T) {
	keyring.MockInit() // empty keyring => Get returns ErrNotFound for this profile
	dir := t.TempDir()
	s := NewStore(dir)
	// Seed only the encrypted-file fallback; keyring has nothing.
	require.NoError(t, s.setFallback("default", "file-only-key"))
	got, err := s.Get("default")
	require.NoError(t, err)
	assert.Equal(t, "file-only-key", got)
}

func TestGetFallback_NoDir(t *testing.T) {
	s := NewStore("") // no fallback dir configured
	_, err := s.getFallback("default")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestSetFallback_NoDirErrors(t *testing.T) {
	s := NewStore("")
	err := s.setFallback("default", "x")
	assert.Error(t, err)
}

func TestDeleteFallback_NoDirNoop(t *testing.T) {
	s := NewStore("")
	assert.NoError(t, s.deleteFallback("default"))
}
