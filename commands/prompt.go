package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// promptSecret prints prompt to stderr and reads one line of input. On a terminal it reads
// without echoing (so a pasted API key never lands in scrollback); on a pipe it falls back to
// a normal line read so scripts still work. It reads a full line — JWT-length keys are fine —
// unlike fmt.Scanln which echoes and mishandles long pastes.
func promptSecret(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		return sanitizeSecret(string(b)), nil
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return sanitizeSecret(line), nil
}

// sanitizeSecret strips terminal bracketed-paste markers (ESC[200~ … ESC[201~) and trims
// surrounding whitespace. When a terminal has bracketed paste enabled, a raw read (unlike the
// shell's line editor) receives those markers wrapping the pasted text; left in, they corrupt a
// pasted API key so it fails auth. Stripping them fixes the common "typing works, pasting 401s".
func sanitizeSecret(s string) string {
	s = strings.ReplaceAll(s, "\x1b[200~", "")
	s = strings.ReplaceAll(s, "\x1b[201~", "")
	return strings.TrimSpace(s)
}
