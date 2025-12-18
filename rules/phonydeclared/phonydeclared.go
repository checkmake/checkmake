// Package phonydeclared implements the ruleset for making sure all targets that don't
// have a rule body are marked PHONY
package phonydeclared

import (
	"fmt"
	"strings"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
)

func init() {
	rules.RegisterRule(&Phonydeclared{})
}

// Phonydeclared is an empty struct on which to call the rule functions
type Phonydeclared struct{}

// Name returns the name of the rule
func (r *Phonydeclared) Name() string {
	return "phonydeclared"
}

// Description returns the description of the rule
func (r *Phonydeclared) Description(cfg rules.RuleConfig) string {
	return "Every target without a body needs to be marked PHONY"
}

// DefaultSeverity returns the default severity for this rule.
func (r *Phonydeclared) DefaultSeverity() rules.Severity {
	return rules.SeverityWarning // Non-phony empty targets can cause incorrect rebuilds
}

// Run executes the rule logic
func (r *Phonydeclared) Run(makefile parser.Makefile, config rules.RuleConfig) rules.RuleViolationList {
	ret := rules.RuleViolationList{}

	ruleIndex := make(map[string]bool)

	// Case 1: .PHONY parsed as variable (old parser behavior)
	for _, variable := range makefile.Variables {
		if variable.Name == "PHONY" {
			for _, phony := range strings.Fields(variable.Assignment) {
				ruleIndex[phony] = true
			}
		}
	}

	// Case 2: .PHONY parsed as rule (new parser behavior)
	for _, rule := range makefile.Rules {
		if rule.Target == ".PHONY" || rule.Target == "PHONY" {
			for _, phony := range rule.Dependencies {
				ruleIndex[phony] = true
			}
		}
	}
	// Check that every non-dot-prefixed target without a body is PHONY
	for _, rule := range makefile.Rules {
		// Skip special or dot-prefixed targets like .PHONY or .DEFAULT_GOAL
		if strings.HasPrefix(rule.Target, ".") {
			continue
		}

		_, ok := ruleIndex[rule.Target]
		if len(rule.Body) == 0 && !ok {
			ret = append(ret, rules.RuleViolation{
				Rule:       "phonydeclared",
				Violation:  fmt.Sprintf("Target %q should be declared PHONY.", rule.Target),
				FileName:   makefile.FileName,
				LineNumber: rule.LineNumber,
			})
		}
	}

	return ret
}
