package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// ref: https://deeeet.com/writing/2014/12/18/golang-cli-test/
func TestRun_versionFlag(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &CLI{OutStream: outStream, ErrStream: errStream}
	args := strings.Split("chfmt -version", " ")

	status := cli.Run(args)
	if status != ExitCodeOK {
		t.Errorf("ExitStatus=%d, want %d", status, ExitCodeOK)
	}

	expected := fmt.Sprintf("chfmt version %s", Version)
	if !strings.Contains(outStream.String(), expected) {
		t.Errorf("Output=%q, want %q", outStream.String(), expected)
	}
}

func TestRun_TooManyArguments(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &CLI{OutStream: outStream, ErrStream: errStream}
	args := strings.Split("chfmt a1 a2 a3", " ")

	status := cli.Run(args)
	if status != ExitCodeTooManyArgs {
		t.Errorf("ExitStatus=%d, want %d", status, ExitCodeTooManyArgs)
	}

	expected := fmt.Sprint("Error: Too many arguments\n")
	if !strings.Contains(errStream.String(), expected) {
		t.Errorf("Output=%q, want %q", outStream.String(), expected)
	}
}
