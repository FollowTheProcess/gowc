// gowc is an experimental reimplementation of the unix coreutils wc tool
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/FollowTheProcess/gowc/internal/count"
)

const usage = `
A toy wc clone (sort of) written in Go.

Usage:
  wc [file...]

Flags:
  -help 	Show this help message and exit
  -json 	Output results as JSON
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
	jsonFlag := flag.Bool("json", false, "Output results as JSON")
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
		result, err := count.One(file, "stdin")
		if err != nil {
			return err
		}
		return result.Display(os.Stdout, *jsonFlag)
	case 1:
		// Read from the file
		path := flag.Arg(0)
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", path, err)
		}
		defer file.Close()
		result, err := count.One(file, path)
		if err != nil {
			return err
		}
		return result.Display(os.Stdout, *jsonFlag)
	default:
		results, err := count.All(flag.Args())
		if err != nil {
			return err
		}

		return results.Display(os.Stdout, *jsonFlag)
	}
}
