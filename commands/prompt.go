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
// and trims surrounding whitespace, unlike fmt.Scanln which echoes and mishandles long pastes.
func promptSecret(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
