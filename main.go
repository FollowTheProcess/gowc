package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"go.followtheprocess.codes/cli"
	"go.followtheprocess.codes/gowc/internal/count"
)

var (
	version = "dev" // gowc version number, set at compile time by ldflags
	commit  = ""    // gowc commit hash, set at compile time with ldflags
	date    = ""    // gowc build date, set at compile time with ldflags
)

func main() {
	if err := run(os.Stdin, os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(stdin io.Reader, stdout, stderr io.Writer, args []string) error {
	var options countOptions
	cmd, err := cli.New(
		"gowc",
		cli.Short("A toy wc clone (sort of) written in Go."),
		cli.Version(version),
		cli.Commit(commit),
		cli.BuildDate(date),
		cli.Example("Use stdin", "cat myfile.txt | gowc"),
		cli.Example("Or", "gowc < myfile.txt"),
		cli.Example("Read from file", "gowc myfile.txt"),
		cli.Example("Or many files", "gowc **/*.go"),
		cli.OverrideArgs(args),
		cli.Stdin(stdin),
		cli.Stdout(stdout),
		cli.Stderr(stderr),
		cli.Flag(&options.json, "json", 'j', false, "Output results as JSON"),
		cli.Allow(cli.AnyArgs()),
		cli.Run(doCount(&options)),
	)
	if err != nil {
		return err
	}

	return cmd.Execute()
}

// countOptions are options passed for a count operation.
type countOptions struct {
	json bool // Format results as JSON
}

func doCount( //nolint: gocognit // This is fine really, it's pretty clear
	options *countOptions,
) func(cmd *cli.Command, args []string) error {
	return func(cmd *cli.Command, args []string) error {
		start := time.Now()
		stdout := cmd.Stdout()
		switch len(args) {
		case 0:
			// Read from stdin
			fd := os.Stdin
			info, err := fd.Stat()
			if err != nil {
				return err
			}
			if info.Size() == 0 {
				return errors.New("nothing to read from stdin")
			}

			result := count.One(fd, "stdin")
			if result.Err != nil {
				return result.Err
			}
			if err := result.Display(stdout, options.json); err != nil {
				return err
			}
		case 1:
			// Read from the file
			path := args[0]
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("could not open %s: %w", path, err)
			}
			defer file.Close()

			result := count.One(file, path)
			if result.Err != nil {
				return result.Err
			}
			if err := result.Display(stdout, options.json); err != nil {
				return err
			}

		default:
			// Count many files
			results := count.All(args)
			for _, result := range results {
				if result.Err != nil {
					return result.Err
				}
			}
			// Sort the results by name so it's deterministic
			slices.SortFunc(results, cmpResult)

			if err := results.Display(stdout, options.json); err != nil {
				return err
			}
		}

		fmt.Fprintf(cmd.Stderr(), "\ntook %v\n", time.Since(start))
		return nil
	}
}

func cmpResult(a, b count.Result) int {
	// Just compare by name
	return cmp.Compare(a.Name, b.Name)
}
