// Package validator holds the basic engine that runs rule based checks
// against a Makefile struct
package validator

import (
	"fmt"

	"github.com/checkmake/checkmake/config"
	"github.com/checkmake/checkmake/logger"
	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"

	// rules register themselves via their package's init function, so we can
	// just blank import it
	_ "github.com/checkmake/checkmake/rules/maxbodylength"
	_ "github.com/checkmake/checkmake/rules/minphony"
	_ "github.com/checkmake/checkmake/rules/phonydeclared"
	_ "github.com/checkmake/checkmake/rules/timestampexpanded"
	_ "github.com/checkmake/checkmake/rules/uniquetargets"
)

// Validate let's you validate a passed in Makefile with the provided config
func Validate(makefile parser.Makefile, cfg *config.Config) (ret rules.RuleViolationList) {
	registeredRules := rules.GetRegisteredRules()

	for name, rule := range registeredRules {
		logger.Debug(fmt.Sprintf("Running rule '%s'...", name))
		ruleConfig := cfg.GetRuleConfig(name)
		if ruleConfig["disabled"] != "true" {
			violations := rule.Run(makefile, ruleConfig)
			// Assign severity to each violation if not already set by the rule
			defaultSeverity := rules.GetSeverityFromConfig(rule, ruleConfig)
			for i := range violations {
				// If severity is not set (empty string), use the default/configured severity
				if violations[i].Severity == "" {
					violations[i].Severity = defaultSeverity
				}
			}
			ret = append(ret, violations...)
		}
	}

	return
}
