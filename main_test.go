package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binName = "gowc"

func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	build := exec.Command("go", "build", `-ldflags=-X 'main.version=0.1.0' -X 'main.commit=blah'`, "-o", binName)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to compile %s: %v\n", binName, err)
		os.Exit(1)
	}

	result := m.Run()

	os.Remove(binName)

	os.Exit(result)
}

func TestHelp(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}

	cmdPath := filepath.Join(dir, binName)

	cmd := exec.Command(cmdPath, "-help")
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("-help produced an error: %v", err)
	}

	want := usage
	got := stdout.String()

	if got != want {
		t.Errorf("\nGot:\t%s\nWanted:\t%s\n", got, want)
	}
}

func TestVersion(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}

	cmdPath := filepath.Join(dir, binName)

	cmd := exec.Command(cmdPath, "-version")
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("-version produced an error: %v", err)
	}

	want := "Version: 0.1.0\nCommit: blah\n"
	got := stdout.String()

	if got != want {
		t.Errorf("\nGot:\n%s\nWanted:\n%s\n", got, want)
	}
}

func TestBadFlag(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cmdPath := filepath.Join(dir, binName)

	cmd := exec.Command(cmdPath, "-bad")
	if err := cmd.Run(); err == nil {
		t.Fatal("-bad did not error")
	}
}

func TestCountStdin(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmdPath := filepath.Join(dir, binName)

	cmd := exec.Command(cmdPath)
	cmd.Stdin = strings.NewReader("hello there\n")
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Reading from stdin returned an error: %s", stderr.String())
	}

	got := stdout.String()

	want := "File\tBytes\tChars\tLines\tWords\nstdin\t12\t12\t1\t2\n"

	if got != want {
		t.Errorf("\nGot:\t%#v\nWanted:\t%#v\n", got, want)
	}
}

func TestCountStdinEmpty(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	stderr := &bytes.Buffer{}

	cmdPath := filepath.Join(dir, binName)

	cmd := exec.Command(cmdPath)
	cmd.Stderr = stderr

	if err := cmd.Run(); err == nil {
		t.Fatalf("Reading from empty stdin did not return an error")
	}

	got := stderr.String()

	want := "nothing to read from stdin\n"

	if got != want {
		t.Errorf("\nGot:\t%#v\nWanted:\t%#v\n", got, want)
	}
}

func TestCountFile(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmdPath := filepath.Join(dir, binName)
	mobyDick := filepath.Join(dir, "internal", "count", "testdata", "moby_dick.txt")

	cmd := exec.Command(cmdPath, mobyDick)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Reading from moby dick returned an error: %s", stderr.String())
	}

	// Windows is stupid and it doesn't even have wc anyway so
	// There's a different number of tabs on windows and the line counts
	// can be different by the looks of it. I don't actually care about windows
	// really so this can go die in a hole
	if runtime.GOOS != "windows" {
		got := stdout.String()

		want := fmt.Sprintf("File\t\t\t\t\t\t\t\t\tBytes\tChars\tLines\tWords\n%s\t1232922\t1232922\t23243\t214132\n", mobyDick)

		if got != want {
			t.Errorf("\nGot:\t%#v\nWanted:\t%#v\n", got, want)
		}
	}
}
