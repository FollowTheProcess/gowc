package main

import (
	"fmt"
	"io"
	"os"

	"github.com/FollowTheProcess/cli"
	"github.com/FollowTheProcess/gowc/internal/count"
)

var (
	version = "dev"     // gowc version number, set at compile time by ldflags
	commit  = "unknown" // gowc commit hash, set at compile time with ldflags
)

func main() {
	if err := run(os.Stdin, os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(stdin io.Reader, stdout io.Writer, stderr io.Writer, args []string) error {
	var options countOptions
	cmd, err := cli.New(
		"gowc",
		cli.Short("A toy wc clone (sort of) written in Go."),
		cli.Version(version),
		cli.Example("Use stdin", "cat myfile.txt | gowc"),
		cli.Example("Or", "gowc < myfile.txt"),
		cli.Example("Read from file", "gowc myfile.txt"),
		cli.Example("Or many files", "gowc **/*.go"),
		cli.Args(args),
		cli.Stdin(stdin),
		cli.Stdout(stdout),
		cli.Stderr(stderr),
		cli.Flag(&options.json, "json", 'j', false, "Output results as JSON"),
		cli.Allow(cli.AnyArgs()),
		cli.Run(doCount(&options)),
		cli.VersionFunc(displayVersion),
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

func doCount(options *countOptions) func(cmd *cli.Command, args []string) error {
	return func(cmd *cli.Command, args []string) error {
		stdout := cmd.Stdout()
		switch len(args) {
		case 0:
			// Read from stdin
			fd := os.Stdin
			info, err := fd.Stat()
			if err != nil {
				return err
			}
			if (info.Mode() & os.ModeCharDevice) != 0 {
				return fmt.Errorf("nothing to read from stdin")
			}

			result, err := count.One(fd, "stdin")
			if err != nil {
				return err
			}
			return result.Display(stdout, options.json)

		case 1:
			// Read from the file
			path := args[0]
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("could not open %s: %w", path, err)
			}
			defer file.Close()

			result, err := count.One(file, path)
			if err != nil {
				return err
			}
			return result.Display(stdout, options.json)

		default:
			// Count many files
			results, err := count.All(args)
			if err != nil {
				return err
			}
			return results.Display(stdout, options.json)
		}
	}
}

func displayVersion(cmd *cli.Command) error {
	fmt.Fprintf(cmd.Stderr(), "Version: %s\nCommit: %s\n", version, commit)
	return nil
}
