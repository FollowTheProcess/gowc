package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"go.followtheprocess.codes/snapshot"
	"go.followtheprocess.codes/test"
)

var (
	update = flag.Bool("update", false, "Update golden files")
	debug  = flag.Bool("debug", false, "Print debug output to stdout")
)

func TestCountFile(t *testing.T) {
	snap := snapshot.New(t, snapshot.Update(*update))

	mobyDick := filepath.Join("internal", "count", "testdata", "TestCount", "moby_dick.txt")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	args := []string{mobyDick}

	err := run(os.Stdin, stdout, stderr, args)
	test.Ok(t, err)

	got := stdout.String()

	if *debug {
		fmt.Printf("\nDEBUG (TestCountFile)\n------------\n\n%s\n", got)
	}

	snap.Snap(got)
}

func TestCountMany(t *testing.T) {
	snap := snapshot.New(t, snapshot.Update(*update))

	args := []string{
		filepath.Join("internal", "count", "testdata", "TestCount", "moby_dick.txt"),
		filepath.Join("internal", "count", "testdata", "TestCount", "another.txt"),
		filepath.Join("internal", "count", "testdata", "TestCount", "onemore.txt"),
		filepath.Join("internal", "count", "testdata", "TestCount", "dir"),
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := run(os.Stdin, stdout, stderr, args)
	test.Ok(t, err)

	got := stdout.String()

	if *debug {
		fmt.Printf("\nDEBUG (TestCountMany)\n------------\n\n%s\n", got)
	}

	snap.Snap(got)
}
