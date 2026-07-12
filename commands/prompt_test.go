package commands

import (
	"strings"
	"testing"
)

func TestScanSecretLine(t *testing.T) {
	// A long paste (well past canonical mode's 1024-char MAX_CANON) reads intact — the whole point.
	long := strings.Repeat("A", 2000)
	if got, err := scanSecretLine(strings.NewReader(long + "\r")); err != nil || got != long {
		t.Errorf("long line: len=%d err=%v", len(got), err)
	}
	// Stops at LF; Backspace (0x7f) edits.
	if got, _ := scanSecretLine(strings.NewReader("ab\x7fc\n")); got != "ac" {
		t.Errorf("backspace: %q", got)
	}
	// Ctrl-C cancels.
	if _, err := scanSecretLine(strings.NewReader("\x03")); err == nil {
		t.Error("Ctrl-C should cancel")
	}
	// EOF with buffered content still returns it.
	if got, _ := scanSecretLine(strings.NewReader("xyz")); got != "xyz" {
		t.Errorf("EOF: %q", got)
	}
}

func TestSanitizeSecretStripsBracketedPaste(t *testing.T) {
	key := "eyJhbGciOiJSUzI1NiJ9.payload.sig"
	// A terminal with bracketed paste on wraps the pasted text in ESC[200~ … ESC[201~; a raw
	// read receives those markers and they must be stripped or the key fails auth.
	if got := sanitizeSecret("\x1b[200~" + key + "\x1b[201~\n"); got != key {
		t.Errorf("bracketed paste not stripped: %q", got)
	}
	// Surrounding whitespace is still trimmed; a clean key is untouched.
	if got := sanitizeSecret("  " + key + "  \r\n"); got != key {
		t.Errorf("trim: %q", got)
	}
	if got := sanitizeSecret(key); got != key {
		t.Errorf("clean key changed: %q", got)
	}
}
