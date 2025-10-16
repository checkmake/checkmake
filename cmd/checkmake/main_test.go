// Package main tests, empty to at least have it be included in the build
package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureOutput temporarily redirects os.Stdout during f(), returning what was printed.
// Useful for testing formatters that write directly to os.Stdout instead of Cobra's writer.
func captureOutput(f func()) string {
	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	f()

	w.Close()
	<-done
	os.Stdout = stdout

	return buf.String()
}

func TestCheckmake_NoArgsShowsHelp(t *testing.T) {
	out := captureOutput(func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{}) // no args
		err := cmd.Execute()
		require.NoError(t, err, "command without args should not fail")
	})

	assert.Contains(t, out, "Usage:", "expected help output to be shown")
	assert.Contains(t, out, "checkmake [flags]", "should display root usage line")
}

func TestCheckmake_RunWithSimpleMakefile(t *testing.T) {
	t.Parallel()
	cmd := newRootCmd()
	cmd.SetArgs([]string{"../../fixtures/simple.make"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.NoError(t, err, "command should run successfully")

	output := buf.String()
	assert.NotContains(t, output, "error", "should not print errors")
}

func TestCheckmake_RunWithViolations(t *testing.T) {
	cmd := newRootCmd()
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"../../fixtures/missing_phony.make"})

	var err error
	out := captureOutput(func() {
		err = cmd.Execute()
	})

	require.Error(t, err, "expected command to fail for a makefile with violations")

	// matching full sentences with table output is complex, search partially
	assert.Contains(t, out, "Missing required", "should mention violation description")
	assert.Contains(t, out, "phony target", "should mention target type")
	assert.Contains(t, out, "phonydeclared", "should mention phonydeclared error")
}

func TestCheckmake_ListRules(t *testing.T) {
	t.Parallel()
	cmd := newRootCmd()
	cmd.SetArgs([]string{"list-rules"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.NoError(t, err, "list-rules should run successfully")

	output := buf.String()
	t.Logf("list-rules output:\n%s", output)

	assert.Regexp(t, `\s+NAME\s+DESCRIPTION\s+`, output)
	assert.Regexp(t, `phonydeclared\s+Every target without a body`, output)
	assert.Regexp(t, `\s+needs\s+to be marked PHONY`, output)
	assert.Regexp(t, `minphony\s+Minimum required phony`, output)
	assert.Regexp(t, `\s+must be present(.*)`, output)
}

func TestCheckmake_WithCustomFormatFlag(t *testing.T) {
	out := captureOutput(func() {
		cmd := newRootCmd()
		cmd.SilenceErrors = true
		cmd.SetArgs([]string{
			"--format", "{{.Rule}} on {{.LineNumber}}",
			"../../fixtures/missing_phony.make",
		})
		_ = cmd.Execute()
	})

	t.Logf("custom format output:\n%s", out)

	require.NotEmpty(t, out, "output should not be empty for custom format")

	// There are three expected rules in missing_phony.make: phonydeclared and minphony twice
	assert.Contains(t, out, "phonydeclared on 16")
	assert.Contains(t, out, "minphony on 21")
}
