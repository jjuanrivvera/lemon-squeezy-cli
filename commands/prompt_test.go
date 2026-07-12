package commands

import "testing"

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
