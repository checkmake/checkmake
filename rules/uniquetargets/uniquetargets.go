// Package uniquetargets implements the ruleset ensuring no target is repeated.
package uniquetargets

import (
	"fmt"
	"strings"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
)

func init() {
	rules.RegisterRule(&UniqueTargets{})
}

// UniqueTargets ensures targets are not defined multiple times.
type UniqueTargets struct{}

// Name returns the rule's identifier.
func (r *UniqueTargets) Name() string {
	return "uniquetargets"
}

// Description returns the human-readable description.
func (r *UniqueTargets) Description(cfg rules.RuleConfig) string {
	if cfg != nil {
		if ignored, ok := cfg["ignore"]; ok && ignored != "" {
			return fmt.Sprintf("Targets should be uniquely defined (ignoring: %s).", ignored)
		}
	}
	return "Targets should be uniquely defined; duplicates can cause recipe overrides or unintended merges."
}

// DefaultSeverity returns the default severity for this rule.
func (r *UniqueTargets) DefaultSeverity() rules.Severity {
	return rules.SeverityError // Duplicate target definitions are errors
}

// Run detects non-unique target definitions, optionally skipping ignored ones.
func (r *UniqueTargets) Run(makefile parser.Makefile, cfg rules.RuleConfig) rules.RuleViolationList {
	seen := make(map[string]int)
	violations := rules.RuleViolationList{}

	// Load optional ignore list
	ignoredTargets := map[string]bool{}
	if cfg != nil {
		if ignoreList, ok := cfg["ignore"]; ok {
			for _, target := range strings.Split(ignoreList, ",") {
				target = strings.TrimSpace(target)
				if target != "" {
					ignoredTargets[target] = true
				}
			}
		}
	}

	for _, rule := range makefile.Rules {
		// Skip .PHONY declarations - they're handled by the multiplephony rule
		if rule.Target == ".PHONY" || rule.Target == "PHONY" {
			continue
		}

		// Skip ignored targets
		if ignoredTargets[rule.Target] {
			continue
		}

		// Skip special built-ins like .PHONY
		if rule.Target == ".PHONY" {
			continue
		}

		if prevLine, exists := seen[rule.Target]; exists {
			violations = append(violations, rules.RuleViolation{
				Rule:       r.Name(),
				Violation:  fmt.Sprintf(`Target "%s" defined multiple times (lines %d and %d).`, rule.Target, prevLine, rule.LineNumber),
				FileName:   makefile.FileName,
				LineNumber: rule.LineNumber,
				Severity:   rules.SeverityError, // Duplicate targets are errors
			})
		} else {
			seen[rule.Target] = rule.LineNumber
		}
	}

	return violations
}
