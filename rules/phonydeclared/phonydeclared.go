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

	// A target with no recipe of its own still builds for real if some
	// other rule supplies the recipe: either an explicit rule for the
	// exact same target elsewhere in the file (a common way to bolt extra
	// prerequisites onto a target, e.g. a hand-written
	// "foo.o: foo.h" alongside the real "foo.o: foo.c\n\t$(CC) ..."
	// recipe, or alongside a pattern rule as below), or a pattern rule
	// ("%.o: %.c") whose stem matches. Neither shape is phony, and
	// flagging them would be actively wrong: PHONY forces the target to
	// rebuild on every invocation, defeating the whole point of a real
	// file target.
	hasRecipeElsewhere := make(map[string]bool)
	var patternRules []parser.Rule
	for _, rule := range makefile.Rules {
		if len(rule.Body) == 0 {
			continue
		}
		if strings.Contains(rule.Target, "%") {
			patternRules = append(patternRules, rule)
		} else {
			hasRecipeElsewhere[rule.Target] = true
		}
	}

	// Check that every non-dot-prefixed target without a body is PHONY
	for _, rule := range makefile.Rules {
		// Skip special or dot-prefixed targets like .PHONY or .DEFAULT_GOAL
		if strings.HasPrefix(rule.Target, ".") {
			continue
		}

		if len(rule.Body) != 0 {
			continue
		}

		if _, ok := ruleIndex[rule.Target]; ok {
			continue
		}

		if hasRecipeElsewhere[rule.Target] || matchesPatternRule(patternRules, rule.Target) {
			continue
		}

		ret = append(ret, rules.RuleViolation{
			Rule:       "phonydeclared",
			Violation:  fmt.Sprintf("Target %q should be declared PHONY.", rule.Target),
			FileName:   makefile.FileName,
			LineNumber: rule.LineNumber,
		})
	}

	return ret
}

// matchesPatternRule reports whether target is covered by one of the given
// pattern rules (targets containing exactly one '%' stem wildcard, GNU
// Make's static/implicit pattern rule syntax, e.g. "%.o: %.c").
func matchesPatternRule(patternRules []parser.Rule, target string) bool {
	for _, p := range patternRules {
		stem := strings.Index(p.Target, "%")
		if stem < 0 {
			continue
		}
		prefix, suffix := p.Target[:stem], p.Target[stem+1:]
		if len(target) >= len(prefix)+len(suffix) &&
			strings.HasPrefix(target, prefix) && strings.HasSuffix(target, suffix) {
			return true
		}
	}
	return false
}
