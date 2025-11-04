// Package main tests, empty to at least have it be included in the build
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
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
	assert.Contains(t, out, "Required target", "should mention violation description")
	assert.Contains(t, out, "declared PHONY.", "should mention target type")
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

func TestCheckmake_DebugLogsMakefilesPassed(t *testing.T) {
	var logBuf bytes.Buffer

	originalLoggerOutput := log.Writer()
	log.SetOutput(&logBuf)

	defer log.SetOutput(originalLoggerOutput)

	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"--debug",
		"../../fixtures/simple.make",
		"../../fixtures/missing_phony.make",
	})
	_ = cmd.Execute()

	logs := logBuf.String()
	t.Logf("debug output:\n%s", logs)

	// The --debug flag should trigger the "Makefiles passed" log.
	require.Contains(t, logs, "Makefiles passed:", "debug output should list the Makefiles provided")

	// And it should contain both files.
	assert.Contains(t, logs, "simple.make")
	assert.Contains(t, logs, "missing_phony.make")
}

func TestCheckmake_ListRules_UsesConfig(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--config", "../../fixtures/custom_rules.ini", "list-rules"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.NoError(t, err, "list-rules should run successfully with config")

	output := buf.String()
	t.Logf("list-rules output:\n%s", output)

	assert.Regexp(t, `3\s+lines`, output, "custom maxBodyLength from config should appear in output")
	assert.Regexp(t, `foo,\s*bar`, output, "custom required phonies from config should appear in output")
}

func TestCheckmake_MinPhonyPassesWhenAllTargetsExist(t *testing.T) {
	cmd := newRootCmd()
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"../../fixtures/all_targets_present.make"})

	var err error
	_ = captureOutput(func() {
		err = cmd.Execute()
	})

	require.NoError(t, err, "expected no violations when all required targets exist and are PHONY")
}

func TestCheckmake_MinPhonyDetectsMissingTargets(t *testing.T) {
	cmd := newRootCmd()
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"../../fixtures/missing_targets.make"})

	var err error
	out := captureOutput(func() {
		err = cmd.Execute()
	})

	require.Error(t, err, "expected command to fail when PHONY declares missing targets")

	assert.Contains(t, out, "Required target", "should mention missing target violation")
	assert.Contains(t, out, "from the Makefile", "should identify missing targets")
	assert.Contains(t, out, "minphony", "should include rule name in output")
}

func TestCheckmake_WithJSONOutput(t *testing.T) {
	out := captureOutput(func() {
		cmd := newRootCmd()
		cmd.SilenceErrors = true
		cmd.SetArgs([]string{
			"-o", "json",
			"../../fixtures/missing_phony.make",
		})
		_ = cmd.Execute()
	})

	t.Logf("JSON output:\n%s", out)

	require.NotEmpty(t, out, "output should not be empty for JSON format")

	// Verify it's valid JSON
	var violations []struct {
		Rule       string `json:"rule"`
		Violation  string `json:"violation"`
		FileName   string `json:"file_name"`
		LineNumber int    `json:"line_number"`
	}

	err := json.Unmarshal([]byte(out), &violations)
	require.NoError(t, err, "output should be valid JSON")

	// Verify we have violations
	assert.Greater(t, len(violations), 0, "should have at least one violation")

	// Verify structure
	for _, v := range violations {
		assert.NotEmpty(t, v.Rule, "rule should not be empty")
		assert.NotEmpty(t, v.Violation, "violation should not be empty")
		assert.NotEmpty(t, v.FileName, "file_name should not be empty")
		assert.Greater(t, v.LineNumber, 0, "line_number should be greater than 0")
	}

	// Verify specific violations are present
	ruleNames := make(map[string]bool)
	for _, v := range violations {
		ruleNames[v.Rule] = true
	}
	assert.Contains(t, ruleNames, "phonydeclared", "should contain phonydeclared violation")
}

func TestCheckmake_WithJSONOutputFlag(t *testing.T) {
	out := captureOutput(func() {
		cmd := newRootCmd()
		cmd.SilenceErrors = true
		cmd.SetArgs([]string{
			"--output", "json",
			"../../fixtures/missing_phony.make",
		})
		_ = cmd.Execute()
	})

	require.NotEmpty(t, out, "output should not be empty for JSON format")

	// Verify it's valid JSON
	var violations []struct {
		Rule       string `json:"rule"`
		Violation  string `json:"violation"`
		FileName   string `json:"file_name"`
		LineNumber int    `json:"line_number"`
	}

	err := json.Unmarshal([]byte(out), &violations)
	require.NoError(t, err, "output should be valid JSON")
}

func TestCheckmake_WithTextOutput(t *testing.T) {
	out := captureOutput(func() {
		cmd := newRootCmd()
		cmd.SilenceErrors = true
		cmd.SetArgs([]string{
			"-o", "text",
			"../../fixtures/missing_phony.make",
		})
		_ = cmd.Execute()
	})

	require.NotEmpty(t, out, "output should not be empty for text format")

	// Text output should contain table-like structure
	assert.Contains(t, out, "Required target", "should mention violation description")
	assert.Contains(t, out, "declared PHONY.", "should mention target type")
	assert.Contains(t, out, "phonydeclared", "should mention phonydeclared error")
}

func TestCheckmake_WithInvalidOutput(t *testing.T) {
	cmd := newRootCmd()
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{
		"-o", "invalid",
		"../../fixtures/missing_phony.make",
	})

	err := cmd.Execute()
	require.Error(t, err, "expected command to fail with invalid output format")
	assert.Contains(t, err.Error(), "invalid output format", "error should mention invalid output format")
}

func TestCheckmake_FormatAndOutputAreMutuallyExclusive(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"--format", "{{.Rule}}: {{.Violation}}",
		"--output", "json",
		"../../fixtures/missing_phony.make",
	})

	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)

	err := cmd.Execute()
	require.Error(t, err, "expected command to fail when both format and output flags are set")

	errorOutput := errBuf.String()
	require.NotEmpty(t, errorOutput, "should produce error output")

	// Cobra handles exclusivity and prints a standard error message.
	// The actual message format:
	// "if any flags in the group [format output] are set none of the others can be; [format output] were all set"
	assert.Contains(t, errorOutput, "[format output]", "should reference the conflicting flags")
	assert.Contains(t, errorOutput, "can be", "should mention that flags cannot be set together")
}
