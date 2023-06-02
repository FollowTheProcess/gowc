package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

const usage = `
A toy wc clone (sort of) written in Go.

Usage:
  wc [file...]

Flags:
  -help 	Show this help message and exit
  -version 	Show version info

Examples:
  # Use stdin
  $ cat myfile.txt | gowc

  # Or
  $ gowc < myfile.txt

  # Read from file
  $ gowc myfile.txt

  # Or many files
  $ gowc **/*.go
`

// TabWriter options.
const (
	minWidth = 0
	tabWidth = 8
	padding  = 1
	padChar  = '\t'
)

type (
	LineCounter int
	ByteCounter int
	WordCounter int
	CharCounter int
)

var (
	version = "dev" // gowc version, set at compile time by ldflags
	commit  = ""    // commit hash at build, set at compile time by ldflags
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	versionFlag := flag.Bool("version", false, "Display version info")
	flag.Usage = func() {
		fmt.Print(usage)
	}

	flag.Parse()

	if *versionFlag {
		fmt.Printf("Version: %s\nCommit: %s\n", version, commit)
		return nil
	}

	switch flag.NArg() {
	case 0:
		// Read from stdin
		file := os.Stdin
		info, err := file.Stat()
		if err != nil {
			return err
		}
		if !((info.Mode() & os.ModeCharDevice) == 0) {
			return fmt.Errorf("nothing to read from stdin")
		}
		return count(file, "stdin")
	case 1:
		// Read from the file
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", flag.Arg(0), err)
		}
		defer file.Close()
		return count(file, flag.Arg(0))
	default:
		// TODO: Support multiple files concurrently
		return errors.New("Multiple files are not supported... yet")
	}
}

func count(in io.Reader, inName string) error {
	var (
		lc LineCounter
		bc ByteCounter
		wc WordCounter
		cc CharCounter
	)

	multi := io.MultiWriter(&lc, &bc, &wc, &cc)

	_, err := io.Copy(multi, in)
	if err != nil {
		return err
	}
	tab := tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, padChar, tabwriter.DiscardEmptyColumns|tabwriter.AlignRight)
	fmt.Fprintln(tab, "File\tBytes\tChars\tLines\tWords")
	fmt.Fprintf(tab, "%s\t%d\t%d\t%d\t%d\n", inName, bc, cc, lc, wc)
	return tab.Flush()
}

// Write implements [io.Writer] for LineCounter.
//
// It doesn't actually write anything, but increments the line count
// on every newline in data, allowing a LineCounter to be used as dst
// in a call to [io.Copy].
func (l *LineCounter) Write(data []byte) (int, error) {
	for _, byt := range data {
		if byt == '\n' {
			*l++
		}
	}
	return len(data), nil
}

// Write implements [io.Writer] for ByteCounter.
//
// It doesn't actually write anything, but increments the byte count
// on every byte in data, allowing a ByteCounter to be used as dst
// in a call to [io.Copy].
func (b *ByteCounter) Write(data []byte) (int, error) {
	for range data {
		*b++
	}
	return len(data), nil
}

// Write implements [io.Writer] for WordCounter.
//
// It doesn't actually write anything, but increments the word count
// on every word in data, allowing a WordCounter to be used as dst
// in a call to [io.Copy].
func (w *WordCounter) Write(data []byte) (int, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		*w++
	}
	return len(data), nil
}

// Write implements [io.Writer] for CharCounter.
//
// It doesn't actually write anything, but increments the char count
// on every utf8 rune in data, allowing a CharCounter to be used as dst
// in a call to [io.Copy].
func (c *CharCounter) Write(data []byte) (int, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanRunes)
	for scanner.Scan() {
		*c++
	}
	return len(data), nil
}
