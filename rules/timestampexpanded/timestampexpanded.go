// Package timestampexpanded implements the ruleset for making sure a variable
// that likely represents a timestamp is simply expanded so it doesn't change
// in between rule executions. Ideally if you want to follow something like
// https://reproducible-builds.org/ timestamps in artefacts are frowned upon,
// however sometimes they are the best tool you have and at least they should
// be consistent across all build artefacts.
package timestampexpanded

import (
	"fmt"
	"strings"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
)

func init() {
	rules.RegisterRule(&Timestampexpanded{})
}

// Timestampexpanded is an empty struct on which to call the rule functions
type Timestampexpanded struct{}

var vT = "Variable %q possibly contains a timestamp and should be simply expanded."

// Name returns the name of the rule
func (r *Timestampexpanded) Name() string {
	return "timestampexpanded"
}

// Description returns the description of the rule
func (r *Timestampexpanded) Description(cfg rules.RuleConfig) string {
	return "timestamp variables should be simply expanded"
}

// DefaultSeverity returns the default severity for this rule.
func (r *Timestampexpanded) DefaultSeverity() rules.Severity {
	return rules.SeverityWarning // Reproducibility issue
}

// Run executes the rule logic
func (r *Timestampexpanded) Run(makefile parser.Makefile, config rules.RuleConfig) rules.RuleViolationList {
	ret := rules.RuleViolationList{}

	for _, variable := range makefile.Variables {
		if strings.Contains(variable.Assignment, " date") &&
			!variable.SimplyExpanded {
			ret = append(ret, rules.RuleViolation{
				Rule:       "timestampexpanded",
				Violation:  fmt.Sprintf(vT, variable.Name),
				FileName:   makefile.FileName,
				LineNumber: variable.LineNumber,
			})
		}
	}

	return ret
}
