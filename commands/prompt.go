package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// promptSecret prints prompt to stderr and reads one line of input. On a terminal it reads
// without echoing (so a pasted API key never lands in scrollback); on a pipe it falls back to a
// normal line read so scripts still work.
//
// The terminal read uses RAW mode, not term.ReadPassword. term.ReadPassword reads in CANONICAL
// mode, whose line buffer is capped at MAX_CANON (1024 bytes on macOS): pasting a longer secret
// (a Lemon Squeezy key is a ~970-char JWT, right at that edge) fills the buffer and the terminal
// BLOCKS further input — the "prompt hangs until Ctrl-C" bug. Raw mode delivers bytes with no
// line-length limit, so long pasted keys read cleanly.
func promptSecret(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		s, err := readSecretRaw(fd)
		fmt.Fprintln(os.Stderr) // raw mode doesn't echo the Enter; end the prompt line
		if err != nil {
			return "", err
		}
		return sanitizeSecret(s), nil
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return sanitizeSecret(line), nil
}

// readSecretRaw reads one line from the terminal in raw, no-echo mode, until CR or LF, with no
// line-length limit (unlike canonical mode's MAX_CANON). Ctrl-C cancels; Backspace/DEL edits.
func readSecretRaw(fd int) (string, error) {
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	var buf []byte
	chunk := make([]byte, 256)
	for {
		n, readErr := os.Stdin.Read(chunk)
		for i := 0; i < n; i++ {
			switch c := chunk[i]; c {
			case '\r', '\n':
				return string(buf), nil
			case 3: // Ctrl-C
				return "", fmt.Errorf("cancelled")
			case 127, 8: // DEL / Backspace
				if len(buf) > 0 {
					buf = buf[:len(buf)-1]
				}
			default:
				buf = append(buf, c)
			}
		}
		if readErr != nil {
			if len(buf) == 0 {
				return "", readErr
			}
			return string(buf), nil
		}
	}
}

// sanitizeSecret strips terminal bracketed-paste markers (ESC[200~ … ESC[201~) and trims
// surrounding whitespace — a defensive guard for terminals that wrap pastes in those markers.
func sanitizeSecret(s string) string {
	s = strings.ReplaceAll(s, "\x1b[200~", "")
	s = strings.ReplaceAll(s, "\x1b[201~", "")
	return strings.TrimSpace(s)
}
