// Package count implements the core counters used in gowc.
package count

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

// TabWriter options.
const (
	minWidth = 0
	tabWidth = 8
	padding  = 1
	padChar  = '\t'
)

type (
	LineCounter uint64
	ByteCounter uint64
	WordCounter uint64
	CharCounter uint64
)

// Result encodes the result of a counting operation on a file.
type Result struct {
	Name  string `json:"name"`
	Lines uint64 `json:"lines"`
	Bytes uint64 `json:"bytes"`
	Words uint64 `json:"words"`
	Chars uint64 `json:"chars"`
}

// Display outputs the Count as a pretty table to w.
func (r Result) Display(w io.Writer, jsonFlag bool) error {
	if jsonFlag {
		if err := json.NewEncoder(w).Encode(r); err != nil {
			return fmt.Errorf("failed to serialise JSON: %w", err)
		}
		return nil
	}
	tab := tabwriter.NewWriter(w, minWidth, tabWidth, padding, padChar, tabwriter.DiscardEmptyColumns|tabwriter.AlignRight)
	fmt.Fprintln(tab, "File\tBytes\tChars\tLines\tWords")
	fmt.Fprintf(tab, "%s\t%d\t%d\t%d\t%d\n", r.Name, r.Bytes, r.Chars, r.Lines, r.Words)
	return tab.Flush()
}

// Results encodes multiple Results from different files.
type Results []Result

// Display outputs the Results to w.
func (r Results) Display(w io.Writer, jsonFlag bool) error {
	if jsonFlag {
		if err := json.NewEncoder(w).Encode(r); err != nil {
			return fmt.Errorf("failed to serialise JSON: %w", err)
		}
		return nil
	}
	tab := tabwriter.NewWriter(w, minWidth, tabWidth, padding, padChar, tabwriter.DiscardEmptyColumns|tabwriter.AlignRight)
	fmt.Fprintln(tab, "File\tBytes\tChars\tLines\tWords")
	for _, res := range r {
		fmt.Fprintf(tab, "%s\t%d\t%d\t%d\t%d\n", res.Name, res.Bytes, res.Chars, res.Lines, res.Words)
	}
	return tab.Flush()
}

// Count performs a counting operation on in, returning the result and any error.
func Count(in io.Reader, name string) (Result, error) {
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
		return Result{}, err
	}

	return Result{
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
