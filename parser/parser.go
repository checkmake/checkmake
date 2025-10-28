// Package parser implements all the parser functionality for Makefiles. This
// is supposed to be a parser with a very small feature set that just supports
// what is needed to do linting and checking and not actual full Makefile
// parsing. And it's handrolled because apparently GNU Make doesn't have a
// grammar (see http://www.mail-archive.com/help-make@gnu.org/msg02778.html)
package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/checkmake/checkmake/logger"
)

// Makefile provides a data structure to describe a parsed Makefile
type Makefile struct {
	FileName  string
	Rules     RuleList
	Variables VariableList
}

// Rule represents a Make rule
type Rule struct {
	Target       string
	Dependencies []string
	Body         []string
	FileName     string
	LineNumber   int
}

// RuleList represents a list of rules
type RuleList []Rule

// Variable represents a Make variable
type Variable struct {
	Name            string
	SimplyExpanded  bool
	Assignment      string
	SpecialVariable bool
	FileName        string
	LineNumber      int
}

// VariableList represents a list of variables
type VariableList []Variable

var (
	// Group 1: The target(s). This is intentionally broad, allowing for special characters (%, .),
	//          variables ($(), ${}), spaces (for multiple targets), and file paths.
	// Group 2: Everything after the colon (prerequisites and/or an inline recipe).
	// Notes: This pattern intentionally excludes variable assignments (":=", "?=", "+=", "!=")
	//        by ensuring that ':' is not immediately followed by '='.
	reFindRule = regexp.MustCompile(`^([A-Za-z0-9_.%/\-$(){}\s]+)\s*:(\s*[^=].*)?$`)

	// reFindRuleBody captures a line belonging to a rule's recipe.
	// It must start with a tab.
	// Group 1: The command to be executed.
	reFindRuleBody = regexp.MustCompile(`^\t+(.*)`)

	// reFindSimpleVariable captures simple/immediate variable assignments.
	// This includes ':=', '::=', and ':::='.
	// Group 1: The variable name (alphanumeric, underscore, dot, hyphen).
	// Group 2: The value being assigned.
	reFindSimpleVariable = regexp.MustCompile(`^([A-Za-z0-9_.-]+)\s*:{1,3}=\s*(.*)`)

	// reFindExpandedVariable captures recursively expanded variable assignments ('=').
	// Group 1: The variable name.
	// Group 2: The value being assigned.
	reFindExpandedVariable = regexp.MustCompile(`^([A-Za-z0-9_.-]+)\s*=\s*(.*)`)

	// reFindOtherVariable captures all other assignment operators.
	// This includes conditional ('?='), shell ('!='), and append ('+=').
	// Group 1: The variable name.
	// Group 2: The operator itself ('?=', '!=', or '+=').
	// Group 3: The value being assigned.
	reFindOtherVariable = regexp.MustCompile(`^([A-Za-z0-9_.-]+)\s*([?!+]=)\s*(.*)`)

	// reFindSpecialTarget captures special Make targets that start with a dot, like .PHONY.
	// Group 1: The special target name (e.g., ".PHONY").
	// Group 2: The prerequisites/dependencies (e.g., "all clean test").
	reFindSpecialTarget = regexp.MustCompile(`^(\.[A-Za-z_]+)\s*:(.*)`)
)

// Parse is the main function to parse a Makefile from a file path string to a
// Makefile struct. This function should be kept fairly small and ideally most
// of the heavy lifting will live in the specific parsing functions below that
// know how to deal with individual lines.
func Parse(filepath string) (ret Makefile, err error) {
	ret.FileName = filepath
	var scanner *MakefileScanner
	scanner, err = NewMakefileScanner(filepath)
	if err != nil {
		return ret, err
	}

	for {
		switch {
		case strings.HasPrefix(scanner.Text(), "#"):
			// parse comments here, ignoring them for now
			scanner.Scan()
		case strings.HasPrefix(scanner.Text(), "."):
			if matches := reFindSpecialTarget.FindStringSubmatch(scanner.Text()); matches != nil {
				// Treat special targets like .PHONY or .DEFAULT_GOAL as rules, not variables
				specialRule := Rule{
					Target:       strings.TrimSpace(matches[1]),
					Dependencies: strings.Fields(strings.TrimSpace(matches[2])),
					Body:         nil,
					FileName:     filepath,
					LineNumber:   scanner.LineNumber,
				}
				ret.Rules = append(ret.Rules, specialRule)
			}
			scanner.Scan()
		default:
			// parse target or variable here, the function advances the scanner
			// itself to be able to detect rule bodies
			ruleOrVariable, parseError := parseRuleOrVariable(scanner)
			if parseError != nil {
				return ret, parseError
			}
			switch v := ruleOrVariable.(type) {
			case Rule:
				ret.Rules = append(ret.Rules, v)
			case Variable:
				ret.Variables = append(ret.Variables, v)
			}

		}

		if scanner.Finished {
			return
		}
	}
}

// parseRuleOrVariable gets the parsing scanner in a state where it resides on
// a line that could be a variable or a rule. The function parses the line and
// subsequent lines if there is a rule body to parse and returns an interface
// that is either a Variable or a Rule struct and leaves the scanner in a
// state where it resides on the first line after the content parsed into the
// returned struct. The parsing of line details is done via regexing for now
// since it seems ok as a first pass but will likely have to change later into
// a proper lexer/parser setup.
//
//nolint:unparam // parseRuleOrVariable never returns an error yet, placeholder for future error handling
func parseRuleOrVariable(scanner *MakefileScanner) (ret interface{}, err error) {
	line := scanner.Text()

	if matches := reFindSimpleVariable.FindStringSubmatch(line); matches != nil {
		ret = Variable{
			Name:           strings.TrimSpace(matches[1]),
			Assignment:     strings.TrimSpace(matches[2]),
			SimplyExpanded: true,
			FileName:       scanner.FileHandle.Name(),
			LineNumber:     scanner.LineNumber,
		}
		scanner.Scan()
		return
	}

	if matches := reFindExpandedVariable.FindStringSubmatch(line); matches != nil {
		ret = Variable{
			Name:           strings.TrimSpace(matches[1]),
			Assignment:     strings.TrimSpace(matches[2]),
			SimplyExpanded: false,
			FileName:       scanner.FileHandle.Name(),
			LineNumber:     scanner.LineNumber,
		}
		scanner.Scan()
		return
	}
	if matches := reFindOtherVariable.FindStringSubmatch(line); matches != nil {
		op := strings.TrimSpace(matches[2])
		isSimple := false // Default to recursive/false

		switch op {
		case "!=":
			// Shell assignment is immediate, like simple expansion.
			isSimple = true
		case "?=":
			// Conditional assignment is recursive, just like '='.
			isSimple = false
		case "+=":
			// Append ('+=') inherits its expansion behavior. If the variable was
			// undefined, '+=' acts like '=' (recursive). Since this parser doesn't
			// track variable history, we default to recursive (false) as the
			// safest and most common-case behavior.
			isSimple = false
		}

		ret = Variable{
			Name:           strings.TrimSpace(matches[1]),
			Assignment:     strings.TrimSpace(matches[3]), // Use index 3 for value
			SimplyExpanded: isSimple,
			FileName:       scanner.FileHandle.Name(),
			LineNumber:     scanner.LineNumber,
		}
		scanner.Scan()
		return
	}

	if matches := reFindRule.FindStringSubmatch(line); matches != nil {
		beginLineNumber := scanner.LineNumber - 1
		scanner.Scan()

		// Handle inline recipe syntax: target: deps ; recipe
		rawDeps := strings.TrimSpace(matches[2])
		inlineRecipe := ""
		if idx := strings.IndexAny(rawDeps, ";"); idx != -1 {
			inlineRecipe = strings.TrimSpace(rawDeps[idx+1:])
			rawDeps = strings.TrimSpace(rawDeps[:idx])
		}

		// Clean up dependencies (space-separated)
		deps := []string{}
		for _, d := range strings.Fields(rawDeps) {
			if d != "" {
				deps = append(deps, d)
			}
		}

		// Collect recipe body (inline + tab-indented)
		ruleBody := []string{}
		if inlineRecipe != "" {
			ruleBody = append(ruleBody, inlineRecipe)
		}

		// collect tab-indented body lines after the rule
		for bodyMatches := reFindRuleBody.FindStringSubmatch(scanner.Text()); bodyMatches != nil; bodyMatches = reFindRuleBody.FindStringSubmatch(scanner.Text()) {
			ruleBody = append(ruleBody, strings.TrimSpace(bodyMatches[1]))
			scanner.Scan()
		}

		ret = Rule{
			Target:       strings.TrimSpace(matches[1]),
			Dependencies: deps,
			Body:         ruleBody,
			FileName:     scanner.FileHandle.Name(),
			LineNumber:   beginLineNumber,
		}
		return
	}

	// Fallback: unrecognized line
	if strings.TrimSpace(line) != "" {
		if strings.Contains(line, ":") && strings.Contains(line, "=") {
			logger.Debug(fmt.Sprintf("Ambiguous line detected: %q (could be variable or rule)", line))
		}
		logger.Debug(fmt.Sprintf("Unable to match line '%s' to a Rule or Variable", line))
	}
	scanner.Scan()
	return
}
