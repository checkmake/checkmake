// Package rules contains specific rules as subpackages to check a Makefile against
package rules

import (
	"sort"

	"github.com/checkmake/checkmake/parser"
)

// Rule defines the interface that all validation rules must implement.
//
// Each rule provides:
//   - Name(): a unique identifier string for the rule.
//   - Description(cfg RuleConfig): a human-readable explanation of what the rule checks for.
//     Implementations should adapt the description based on the provided configuration,
//     but must remain safe to call with a nil config (using default values).
//   - DefaultSeverity(): returns the default severity level for violations from this rule.
//   - Run(makefile, cfg): performs the actual validation on the parsed Makefile,
//     returning a list of any violations found.
type Rule interface {
	Name() string
	Description(cfg RuleConfig) string
	DefaultSeverity() Severity
	Run(parser.Makefile, RuleConfig) RuleViolationList
}

// Severity represents the severity level of a rule violation
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// RuleViolation represents a basic validation failure
type RuleViolation struct {
	Rule       string
	Violation  string
	FileName   string
	LineNumber int
	Severity   Severity
}

// RuleViolationList is a list of Violation types and the return type of a
// Rule function
type RuleViolationList []RuleViolation

// RuleConfig is a simple string/string map to hold key/value configuration
// for rules.
type RuleConfig map[string]string

// RuleConfigMap is a map that stores RuleConfig maps keyed by the rule name
type RuleConfigMap map[string]RuleConfig

// RuleRegistry is the type to hold rules keyed by their name
type RuleRegistry map[string]Rule

var ruleRegistry RuleRegistry

func init() {
	ruleRegistry = make(RuleRegistry)
}

// RegisterRule let's you register a rule for inclusion in the validator
func RegisterRule(r Rule) {
	ruleRegistry[r.Name()] = r
}

// GetRegisteredRules returns the internal ruleRegistry
func GetRegisteredRules() RuleRegistry {
	return ruleRegistry
}

// GetRulesSorted returns all registered rules in alphabetical order by name.
func GetRulesSorted() []Rule {
	keys := make([]string, 0, len(ruleRegistry))
	for name := range ruleRegistry {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	sorted := make([]Rule, 0, len(keys))
	for _, name := range keys {
		sorted = append(sorted, ruleRegistry[name])
	}
	return sorted
}

// GetSeverityFromConfig returns the severity for a rule from its config, or the default if not specified.
func GetSeverityFromConfig(rule Rule, cfg RuleConfig) Severity {
	if cfg != nil {
		if severityStr, ok := cfg["severity"]; ok {
			severity := Severity(severityStr)
			// Validate severity value
			if severity == SeverityError || severity == SeverityWarning || severity == SeverityInfo {
				return severity
			}
		}
	}
	return rule.DefaultSeverity()
}
