package formatters

import (
	"io"
	"log"
	"os"
	"strconv"

	"github.com/checkmake/checkmake/rules"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// DefaultFormatter is the formatter used by default for CLI output
type DefaultFormatter struct {
	out io.Writer
}

// NewDefaultFormatter returns a DefaultFormatter struct
func NewDefaultFormatter() *DefaultFormatter {
	return &DefaultFormatter{out: os.Stdout}
}

// Format is the function to call to get the formatted output
func (f *DefaultFormatter) Format(violations rules.RuleViolationList) {
	data := make([][]string, len(violations))

	for idx, val := range violations {
		severityStr := string(val.Severity)
		if severityStr == "" {
			severityStr = "error" // Default to error for backward compatibility
		}
		data[idx] = []string{
			severityStr,
			val.Rule,
			val.Violation,
			val.FileName,
			strconv.Itoa(val.LineNumber),
		}
	}

	table := tablewriter.NewTable(f.out,
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.BorderNone,
			Symbols: tw.NewSymbols(tw.StyleNone),
			Settings: tw.Settings{
				Lines:      tw.LinesNone,
				Separators: tw.SeparatorsNone,
			},
		}),
		tablewriter.WithRowAutoWrap(tw.WrapNormal),
		tablewriter.WithMaxWidth(80),
	)

	table.Header("Severity", "Rule", "Description", "File Name", "Line Number")

	if err := table.Bulk(data); err != nil {
		log.Fatalf("Bulk append failed: %v", err)
	}
	table.Render()
}
