package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	snapshot := flag.Bool("snapshot", false, "нарисовать один кадр и выйти (для превью)")
	snapW := flag.Int("w", 100, "ширина кадра для -snapshot")
	snapH := flag.Int("h", 28, "высота кадра для -snapshot")
	screen := flag.String("screen", "play", "кадр для -snapshot: play, start или over")
	flag.Parse()

	if *snapshot {
		runSnapshot(*snapW, *snapH, *screen)
		return
	}

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

func run() error {
	term, err := setupTerminal()
	if err != nil {
		return fmt.Errorf("%w\nЗапустите игру в обычном терминале: go run .", err)
	}
	defer term.restore()

	w, h := terminalSize()
	g := newGame(w, h, loadBest())
	g.persist = true

	keys := make(chan key, 32)
	go readKeys(keys)

	sigCh := make(chan os.Signal, 4)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGWINCH)
	defer signal.Stop(sigCh)

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	last := time.Now()
	var frame bytes.Buffer
	out := bufio.NewWriterSize(os.Stdout, 1<<16)

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			dt := now.Sub(last).Seconds()
			last = now

			if !g.tooSmall() {
				g.update(dt)
			}

			frame.Reset()
			frame.WriteString(g.renderFrame())
			if _, err := out.Write(frame.Bytes()); err != nil {
				return err
			}
			if err := out.Flush(); err != nil {
				return err
			}

		case k, ok := <-keys:
			if !ok {
				return nil
			}
			if g.handleKey(k) {
				return nil
			}

		case s := <-sigCh:
			if s == syscall.SIGWINCH {
				nw, nh := terminalSize()
				g.resize(nw, nh)
				last = time.Now()
				continue
			}
			return nil
		}
	}
}

func runSnapshot(w, h int, screen string) {
	if w < 40 {
		w = 40
	}
	if h < 16 {
		h = 16
	}
	fmt.Println(snapshotFrame(w, h, screen))
}
