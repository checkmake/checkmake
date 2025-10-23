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

// Run executes the rule logic
func (r *Phonydeclared) Run(makefile parser.Makefile, config rules.RuleConfig) rules.RuleViolationList {
	ret := rules.RuleViolationList{}

	ruleIndex := make(map[string]bool)

	for _, variable := range makefile.Variables {
		if variable.Name == "PHONY" {
			for _, phony := range strings.Split(variable.Assignment, " ") {
				ruleIndex[phony] = true
			}
		}
	}

	for _, rule := range makefile.Rules {
		_, ok := ruleIndex[rule.Target]
		if len(rule.Body) == 0 && ok == false {
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
