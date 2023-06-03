// gowc is an experimental reimplementation of the unix coreutils wc tool
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

var (
	version = "dev" // gowc version, set at compile time by ldflags
	commit  = ""    // commit hash at build, set at compile time by ldflags
)

type (
	LineCounter uint64
	ByteCounter uint64
	WordCounter uint64
	CharCounter uint64
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// TODO: Add a -json flag
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
		result, err := count(file, "stdin")
		if err != nil {
			return err
		}
		return result.Display(os.Stdout)
	case 1:
		// Read from the file
		path := flag.Arg(0)
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", path, err)
		}
		defer file.Close()
		result, err := count(file, path)
		if err != nil {
			return err
		}
		return result.Display(os.Stdout)
	default:
		// TODO: Support multiple files concurrently
		return errors.New("Multiple files are not supported... yet")
	}
}

// count performs a counting operation on in, returning the result and any error.
func count(in io.Reader, name string) (Count, error) {
	var (
		lc LineCounter
		bc ByteCounter
		wc WordCounter
		cc CharCounter
	)

	// TODO: This currently copies to each writer one at a time when there's really no need
	// they are all separate so could be parallelised
	multi := io.MultiWriter(&lc, &bc, &wc, &cc)

	_, err := io.Copy(multi, in)
	if err != nil {
		return Count{}, err
	}

	return Count{
		Name:  name,
		Lines: uint64(lc),
		Bytes: uint64(bc),
		Words: uint64(wc),
		Chars: uint64(cc),
	}, nil
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

// Count encodes the result of a counting operation on a file.
type Count struct {
	Name  string
	Lines uint64
	Bytes uint64
	Words uint64
	Chars uint64
}

// Display outputs the Count as a pretty table to w.
func (c Count) Display(w io.Writer) error {
	tab := tabwriter.NewWriter(w, minWidth, tabWidth, padding, padChar, tabwriter.DiscardEmptyColumns|tabwriter.AlignRight)
	fmt.Fprintln(tab, "File\tBytes\tChars\tLines\tWords")
	fmt.Fprintf(tab, "%s\t%d\t%d\t%d\t%d\n", c.Name, c.Bytes, c.Chars, c.Lines, c.Words)
	return tab.Flush()
}
