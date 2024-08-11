package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FollowTheProcess/test"
)

const binName = "gowc"

var binPath string

func TestMain(m *testing.M) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get cwd: %v", err)
		os.Exit(1)
	}
	bin := filepath.Join(cwd, binName)
	build := exec.Command("go", "build", `-ldflags=-X 'main.version=0.1.0' -X 'main.commit=blah'`, "-o", bin)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to compile %s: %v\n", bin, err)
		os.Exit(1)
	}

	binPath = bin

	result := m.Run()

	os.Remove(binPath)

	os.Exit(result)
}

func TestHelp(t *testing.T) {
	stdout := &bytes.Buffer{}

	cmd := exec.Command(binPath, "-help")
	cmd.Stdout = stdout
	err := cmd.Run()
	test.Ok(t, err)

	want := usage
	got := stdout.String()

	test.Equal(t, got, want)
}

func TestVersion(t *testing.T) {
	stdout := &bytes.Buffer{}

	cmd := exec.Command(binPath, "-version")
	cmd.Stdout = stdout
	err := cmd.Run()
	test.Ok(t, err)

	want := "Version: 0.1.0\nCommit: blah\n"
	got := stdout.String()

	test.Equal(t, got, want)
}

func TestBadFlag(t *testing.T) {
	cmd := exec.Command(binPath, "-bad")
	if err := cmd.Run(); err == nil {
		t.Fatal("-bad did not error")
	}
}

func TestCountStdin(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command(binPath)
	cmd.Stdin = strings.NewReader("hello there\n")
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	test.Ok(t, err)

	got := stdout.String()

	want := "File\tBytes\tChars\tLines\tWords\nstdin\t12\t12\t1\t2\n"

	test.Equal(t, got, want)
}

func TestCountStdinJSON(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command(binPath, "-json")
	cmd.Stdin = strings.NewReader("hello there\n")
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	test.Ok(t, err)

	got := strings.TrimSpace(stdout.String())

	want := `{"name":"stdin","lines":1,"bytes":12,"words":2,"chars":12}`

	test.Equal(t, got, want)
}

func TestCountStdinEmpty(t *testing.T) {
	stderr := &bytes.Buffer{}

	cmd := exec.Command(binPath)
	cmd.Stderr = stderr

	if err := cmd.Run(); err == nil {
		t.Fatalf("Reading from empty stdin did not return an error")
	}

	got := stderr.String()

	want := "nothing to read from stdin\n"

	test.Equal(t, got, want)
}

func TestCountFile(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	mobyDick := filepath.Join("internal", "count", "testdata", "moby_dick.txt")

	cmd := exec.Command(binPath, mobyDick)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	test.Ok(t, err)

	got := stdout.String()

	want := fmt.Sprintf("File\t\t\t\t\tBytes\tChars\tLines\tWords\n%s\t1232922\t1232922\t23243\t214132\n", mobyDick)

	test.Equal(t, got, want)
}

func TestCountFileJSON(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	mobyDick := filepath.Join("internal", "count", "testdata", "moby_dick.txt")

	cmd := exec.Command(binPath, "-json", mobyDick)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	test.Ok(t, err)

	got := strings.TrimSpace(stdout.String())

	want := fmt.Sprintf(`{"name":"%s","lines":23243,"bytes":1232922,"words":214132,"chars":1232922}`, mobyDick)

	test.Equal(t, got, want)
}
