package formatters

import (
	"encoding/json"
	"io"
	"os"

	"github.com/checkmake/checkmake/rules"
)

// JSONFormatter is the formatter used for JSON output
type JSONFormatter struct {
	out io.Writer
}

// NewJSONFormatter returns a JSONFormatter struct
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{out: os.Stdout}
}

// Format is the function to call to get the formatted JSON output
func (f *JSONFormatter) Format(violations rules.RuleViolationList) {
	// Convert violations to JSON-serializable structure
	type ViolationJSON struct {
		Rule       string `json:"rule"`
		Violation  string `json:"violation"`
		FileName   string `json:"file_name"`
		LineNumber int    `json:"line_number"`
	}

	violationsJSON := make([]ViolationJSON, len(violations))
	for i, v := range violations {
		violationsJSON[i] = ViolationJSON{
			Rule:       v.Rule,
			Violation:  v.Violation,
			FileName:   v.FileName,
			LineNumber: v.LineNumber,
		}
	}

	encoder := json.NewEncoder(f.out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(violationsJSON); err != nil {
		// If encoding fails, we can't really recover, so we'll just return
		// The error will be visible in the output stream
		return
	}
}
