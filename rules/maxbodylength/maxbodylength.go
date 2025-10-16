// Package maxbodylength implements the ruleset for making sure target bodies
// are kept short and thus hopefully somewhat not complex.
package maxbodylength

import (
	"fmt"
	"strconv"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
)

var (
	maxBodyLength = 5
)

func init() {
	rules.RegisterRule(&MaxBodyLength{})
}

// MaxBodyLength is an empty struct on which to call the rule functions
type MaxBodyLength struct {
}

var (
	vT = "Target body for %q exceeds allowed length of %d lines (%d)."
)

// Name returns the name of the rule
func (m *MaxBodyLength) Name() string {
	return "maxbodylength"
}

// Description returns the description of the rule
func (m *MaxBodyLength) Description(cfg rules.RuleConfig) string {
	if cfg != nil {
		if confLength, ok := cfg["maxBodyLength"]; ok {
			if i, err := strconv.Atoi(confLength); err == nil {
				return fmt.Sprintf("Target bodies should be kept simple and short (no more than %d lines).", i)
			}
		}
	}
	return fmt.Sprintf("Target bodies should be kept simple and short (no more than %d lines).", maxBodyLength)
}

// Run executes the rule logic
func (m *MaxBodyLength) Run(makefile parser.Makefile, config rules.RuleConfig) rules.RuleViolationList {
	ret := rules.RuleViolationList{}

	if confLength, ok := config["maxBodyLength"]; ok {
		if i, err := strconv.Atoi(confLength); err == nil {
			maxBodyLength = i
		}
	}

	for _, rule := range makefile.Rules {
		if len(rule.Body) > maxBodyLength {
			ret = append(ret, rules.RuleViolation{
				Rule:       "maxbodylength",
				Violation:  fmt.Sprintf(vT, rule.Target, maxBodyLength, len(rule.Body)),
				FileName:   makefile.FileName,
				LineNumber: rule.LineNumber,
			})
		}
	}

	return ret
}
