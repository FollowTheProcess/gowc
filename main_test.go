package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
