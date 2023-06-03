package main

import (
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

	build := exec.Command("go", "build", "-o", binName)
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

	cmdPath := filepath.Join(dir, binName)

	cmd := exec.Command(cmdPath, "-help")
	if err := cmd.Run(); err != nil {
		t.Fatalf("-help produced an error: %v", err)
	}
}

func TestVersion(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cmdPath := filepath.Join(dir, binName)

	cmd := exec.Command(cmdPath, "-version")
	if err := cmd.Run(); err != nil {
		t.Fatalf("-version produced an error: %v", err)
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
