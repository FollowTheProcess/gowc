// Package count implements the core counters used in gowc.
package count

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
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
	Name  string      `json:"name"`
	Lines LineCounter `json:"lines"`
	Bytes ByteCounter `json:"bytes"`
	Words WordCounter `json:"words"`
	Chars CharCounter `json:"chars"`
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

// One performs a counting operation on a single reader, returning the result and any error.
func One(in io.Reader, name string) (Result, error) {
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
		Lines: lc,
		Bytes: bc,
		Words: wc,
		Chars: cc,
	}, nil
}

// All performs counting operations on multiple files concurrently
// gathering up the results and returning.
func All(files []string) (Results, error) {
	jobs := make(chan string)
	counts := make(chan Result)

	// Keep a waitgroup so we know when all the workers are done
	var wg sync.WaitGroup

	// Launch a concurrent worker pool to chew through the queue of files to hash
	// these will all initially block as no files are on the jobs channel yet
	// nWorkers is min of NumCPU and len(files) so we don't start more workers than
	// is necessary (no point kicking off 8 workers to do 3 files for example)
	nWorkers := min(runtime.NumCPU(), len(files))
	for range nWorkers {
		wg.Add(1)
		go worker(counts, jobs, &wg)
	}

	// Load files onto the jobs channel, this is a goroutine so it
	// doesnt block the main goroutine as channel cap is 0
	go func() {
		for _, file := range files {
			jobs <- file
		}
		close(jobs)
	}()

	// Wait for all the workers to finish in another goroutine, again so it
	// doesn't block the main routine, and close counts channel when done
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(counts)
	}(&wg)

	results := make(Results, 0, len(files))

	// Finally, range over the counts channel until it gets closed by the
	// goroutine above, adding each result to the list of results to be returned
	for count := range counts {
		results = append(results, count)
	}

	return results, nil
}

// worker is a concurrent worker contributing to counting in files,
// it pulls files off the jobs channel, counts things in them, and puts
// it's results on the counts channel. It takes a reference to a WaitGroup
// so we can be sure all workers have finished before closing the counts channel.
func worker(counts chan<- Result, files <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for file := range files {
		f, err := os.Open(file)
		if err != nil {
			// If we can't open the (possibly dir) path
			// then just close the file and continue
			f.Close()
			continue
		}
		info, _ := f.Stat() //nolint: errcheck // The file is already open here so we can ignore the error
		// Skip directories
		if info.IsDir() {
			// Same as above, ensure we close the file before carrying on
			f.Close()
			continue
		}
		result, err := One(f, file)
		if err != nil {
			// Same as above
			f.Close()
			continue
		}
		f.Close()
		counts <- result
	}
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
