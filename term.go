package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode/utf8"
)

type key int

const (
	keyJump key = iota
	keyQuit
	keyPause
	keyRestart
)

type terminal struct {
	saved string
}

func setupTerminal() (*terminal, error) {
	if _, err := exec.LookPath("stty"); err != nil {
		return nil, errors.New("не найдена утилита stty")
	}
	saved, err := stty("-g")
	if err != nil || saved == "" {
		return nil, errors.New("нужен интерактивный терминал")
	}
	if _, err := stty("-echo", "-icanon", "min", "1", "time", "0"); err != nil {
		_, _ = stty(saved)
		return nil, fmt.Errorf("не удалось перевести терминал в игровой режим: %w", err)
	}

	fmt.Print("\x1b[?1049h\x1b[?25l\x1b[?7l\x1b[2J")
	return &terminal{saved: saved}, nil
}

func (t *terminal) restore() {
	fmt.Print("\x1b[0m\x1b[?7h\x1b[?25h\x1b[?1049l")
	if t.saved != "" {
		_, _ = stty(strings.Fields(t.saved)...)
	}
}

func stty(args ...string) (string, error) {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = os.Stdin
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

func terminalSize() (int, int) {
	if out, err := stty("size"); err == nil {
		var h, w int
		if _, scanErr := fmt.Sscanf(out, "%d %d", &h, &w); scanErr == nil && w > 0 && h > 0 {
			return w, h
		}
	}
	w, _ := strconv.Atoi(os.Getenv("COLUMNS"))
	h, _ := strconv.Atoi(os.Getenv("LINES"))
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	return w, h
}

func readKeys(ch chan<- key) {
	buf := make([]byte, 128)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			close(ch)
			return
		}
		for i := 0; i < n; {
			b := buf[i]

			// Arrow keys: ESC [ A/B/C/D.
			if b == 0x1b {
				if i+2 < n && buf[i+1] == '[' {
					if buf[i+2] == 'A' {
						sendKey(ch, keyJump)
					}
					i += 3
					continue
				}
				i++
				continue
			}

			if b < utf8.RuneSelf {
				switch b {
				case ' ', 'w', 'W', 'j', 'J', '\r', '\n':
					sendKey(ch, keyJump)
				case 'q', 'Q', 3:
					sendKey(ch, keyQuit)
				case 'p', 'P':
					sendKey(ch, keyPause)
				case 'r', 'R':
					sendKey(ch, keyRestart)
				}
				i++
				continue
			}

			r, size := utf8.DecodeRune(buf[i:n])
			if r == utf8.RuneError && size == 1 {
				i++
				continue
			}
			switch r {
			case 'ц', 'Ц':
				sendKey(ch, keyJump)
			case 'й', 'Й':
				sendKey(ch, keyQuit)
			case 'з', 'З':
				sendKey(ch, keyPause)
			case 'к', 'К':
				sendKey(ch, keyRestart)
			}
			i += size
		}
	}
}

func sendKey(ch chan<- key, k key) {
	select {
	case ch <- k:
	default:
	}
}
