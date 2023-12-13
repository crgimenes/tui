package main

import (
	"log"
	"os"
	"os/signal"

	"golang.org/x/term"
)

type termIO struct {
	width  int
	height int
}

func (t *termIO) Read(p []byte) (n int, err error) {
	return os.Stdin.Read(p)
}

func (t *termIO) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (t *termIO) SetSize(width, height int) {
	t.width = width
	t.height = height
}

func main() {
	if !term.IsTerminal(0) {
		println("not in a term")
		return
	}

	width, height, err := term.GetSize(0)
	if err != nil {
		println("error getting term size")
		return
	}

	termIO := &termIO{}
	termIO.SetSize(width, height)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
		println("\033[?25h")
	}()

	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, os.Interrupt)
		<-sc

		_ = term.Restore(int(os.Stdin.Fd()), oldState)

		log.Println("shutting down...")

		os.Exit(0)
	}()

	// clear screen
	termIO.Write([]byte("\033[2J"))

	go func() {
		for {
			buffer := make([]byte, 256*1024)
			n, err := termIO.Read(buffer)
			if err != nil {
				panic(err)
			}

			if buffer[0] == '\x03' {
				_ = term.Restore(int(os.Stdin.Fd()), oldState)
				println("\033[?25h")
				os.Exit(0)
			}

			termIO.Write(buffer[:n])
		}
	}()

	<-make(chan struct{})

}
