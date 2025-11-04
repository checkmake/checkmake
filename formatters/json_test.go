package formatters

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/checkmake/checkmake/config"
	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
	"github.com/checkmake/checkmake/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONFormatter(t *testing.T) {
	out := new(bytes.Buffer)
	formatter := JSONFormatter{out: out}

	makefile, _ := parser.Parse("../fixtures/missing_phony.make")

	violations := validator.Validate(makefile, &config.Config{})
	formatter.Format(violations)

	// Verify JSON output
	var violationsJSON []struct {
		Rule       string `json:"rule"`
		Violation  string `json:"violation"`
		FileName   string `json:"file_name"`
		LineNumber int    `json:"line_number"`
		Severity   string `json:"severity"`
	}

	err := json.Unmarshal(out.Bytes(), &violationsJSON)
	require.NoError(t, err, "output should be valid JSON")

	// Verify we have violations
	assert.Greater(t, len(violationsJSON), 0, "should have at least one violation")

	// Verify structure
	for _, v := range violationsJSON {
		assert.NotEmpty(t, v.Rule, "rule should not be empty")
		assert.NotEmpty(t, v.Violation, "violation should not be empty")
		assert.NotEmpty(t, v.FileName, "file_name should not be empty")
		assert.Greater(t, v.LineNumber, 0, "line_number should be greater than 0")
		assert.NotEmpty(t, v.Severity, "severity should not be empty")
		assert.Contains(t, []string{"error", "warning", "info"}, v.Severity, "severity should be one of error, warning, or info")
	}

	// Verify specific violations are present
	ruleNames := make(map[string]bool)
	for _, v := range violationsJSON {
		ruleNames[v.Rule] = true
	}
	assert.Contains(t, ruleNames, "phonydeclared", "should contain phonydeclared violation")
}

func TestJSONFormatter_EmptyViolations(t *testing.T) {
	out := new(bytes.Buffer)
	formatter := JSONFormatter{out: out}

	var violations []struct {
		Rule       string `json:"rule"`
		Violation  string `json:"violation"`
		FileName   string `json:"file_name"`
		LineNumber int    `json:"line_number"`
		Severity   string `json:"severity"`
	}

	formatter.Format(rules.RuleViolationList{})

	err := json.Unmarshal(out.Bytes(), &violations)
	require.NoError(t, err, "output should be valid JSON even with no violations")
	assert.Equal(t, 0, len(violations), "should have empty array for no violations")
}
