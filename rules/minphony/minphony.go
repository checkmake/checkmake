// Package minphony implements the ruleset for making sure required minimum
// phony targets are present
package minphony

import (
	"fmt"
	"strings"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
)

var (
	defaultRequired = []string{
		"all",
		"clean",
		"test",
	}
)

func init() {
	rules.RegisterRule(&MinPhony{required: defaultRequired})
}

// MinPhony is an empty struct on which to call the rule functions
type MinPhony struct {
	required []string
}

// Name returns the name of the rule
func (r *MinPhony) Name() string {
	return "minphony"
}

// Description returns the description of the rule
func (r *MinPhony) Description(cfg rules.RuleConfig) string {
	if cfg != nil {
		if req, ok := cfg["required"]; ok && req != "" {
			return fmt.Sprintf("Minimum required phony targets must be present (%s).", req)
		}
	}
	return fmt.Sprintf("Minimum required phony targets must be present (%s).", strings.Join(r.required, ","))
}

// Run executes the rule logic
func (r *MinPhony) Run(makefile parser.Makefile, config rules.RuleConfig) rules.RuleViolationList {
	ret := rules.RuleViolationList{}

	if confRequired, ok := config["required"]; ok {
		// special case:
		// empty string means disable the rule.
		if confRequired == "" {
			r.required = []string{}
		} else {

			r.required = strings.Split(confRequired, ",")
		}
		for i := range r.required {
			r.required[i] = strings.TrimSpace(r.required[i])
		}

	}
	ruleIndex := make(map[string]bool)
	ruleLineNumber := 0
	for _, variable := range makefile.Variables {
		if variable.Name == "PHONY" {
			ruleLineNumber = variable.LineNumber - 1
			for _, phony := range strings.Split(variable.Assignment, " ") {
				ruleIndex[phony] = true
			}
		}
	}

	for _, reqRule := range r.required {
		_, ok := ruleIndex[reqRule]
		if !ok {
			ret = append(ret, rules.RuleViolation{
				Rule:       "minphony",
				Violation:  fmt.Sprintf("Missing required phony target %q", reqRule),
				FileName:   makefile.FileName,
				LineNumber: ruleLineNumber,
			})
		}
	}

	return ret
}
