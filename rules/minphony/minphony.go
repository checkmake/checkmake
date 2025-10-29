// Package minphony implements the ruleset for making sure required minimum
// phony targets are present
package minphony

import (
	"fmt"
	"strings"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
)

var defaultRequired = []string{
	"all",
	"clean",
	"test",
}

func init() {
	rules.RegisterRule(&MinPhony{required: defaultRequired})
}

//MinPhony is an empty struct on which to call the rule functions
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

// Run executes the rule logic.
// It ensures all required phony targets are both defined as rules
// and declared as PHONY. Missing or undeclared targets trigger violations.
func (r *MinPhony) Run(makefile parser.Makefile, config rules.RuleConfig) rules.RuleViolationList {
	ret := rules.RuleViolationList{}

	// Load configured required targets, if any
	required := r.required
	if confRequired, ok := config["required"]; ok {
		if confRequired == "" {
			required = []string{}
		} else {
			required = strings.Split(confRequired, ",")
		}
		for i := range required {
			required[i] = strings.TrimSpace(required[i])
		}
	}

	// Collect all declared phony targets
	declaredPhony := map[string]bool{}
	phonyLine := 0

	// .PHONY parsed as variable (old behavior)
	for _, variable := range makefile.Variables {
		if variable.Name == "PHONY" {
			phonyLine = variable.LineNumber
			for _, phony := range strings.Fields(variable.Assignment) {
				declaredPhony[phony] = true
			}
		}
	}

	// .PHONY parsed as rule (new parser behavior)
	for _, rule := range makefile.Rules {
		if rule.Target == ".PHONY" || rule.Target == "PHONY" {
			phonyLine = rule.LineNumber
			for _, phony := range rule.Dependencies {
				declaredPhony[phony] = true
			}
		}
	}

	// NOTE: historically, when .PHONY was parsed as a variable,
	// the reported line number was adjusted with `-1`. Now that
	// .PHONY is parsed as a rule, we use the accurate line number.
	// This change causes a +1 shift in violation line reporting.
	//
	// Fallback: ensure phonyLine is never undefined
	if phonyLine == 0 {
		if len(makefile.Rules) > 0 {
			phonyLine = makefile.Rules[len(makefile.Rules)-1].LineNumber
		}
		if phonyLine == 0 {
			phonyLine = -1 // match historical behavior for missing PHONY line
		}
	}

	// Collect all defined targets in the Makefile
	definedTargets := map[string]bool{}
	for _, rule := range makefile.Rules {
		definedTargets[rule.Target] = true
	}

	// Check for required targets being both defined and declared PHONY
	for _, req := range required {
		// Check if the required target is defined at all
		if !definedTargets[req] {
			ret = append(ret, rules.RuleViolation{
				Rule:       r.Name(),
				Violation:  fmt.Sprintf("Required target %q is missing from the Makefile.", req),
				FileName:   makefile.FileName,
				LineNumber: phonyLine,
			})
			continue
		}

		// Check if it’s declared PHONY
		if !declaredPhony[req] {
			ret = append(ret, rules.RuleViolation{
				Rule:       r.Name(),
				Violation:  fmt.Sprintf("Required target %q must be declared PHONY.", req),
				FileName:   makefile.FileName,
				LineNumber: phonyLine,
			})
		}
	}

	return ret
}
