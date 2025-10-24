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
	reFindRule             = regexp.MustCompile(`^([a-zA-Z]+):(.*)`)
	reFindRuleBody         = regexp.MustCompile(`^\t+(.*)`)
	reFindSimpleVariable   = regexp.MustCompile(`^([a-zA-Z]+) ?:=(.*)`)
	reFindExpandedVariable = regexp.MustCompile(`^([a-zA-Z]+) ?=(.*)`)
	reFindSpecialVariable  = regexp.MustCompile(`^\.([a-zA-Z_]+):(.*)`)
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
			if matches := reFindSpecialVariable.FindStringSubmatch(scanner.Text()); matches != nil {
				specialVar := Variable{
					Name:            strings.TrimSpace(matches[1]),
					Assignment:      strings.TrimSpace(matches[2]),
					SpecialVariable: true,
					FileName:        filepath,
					LineNumber:      scanner.LineNumber,
				}
				ret.Variables = append(ret.Variables, specialVar)
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
func parseRuleOrVariable(scanner *MakefileScanner) (ret interface{}, err error) {
	line := scanner.Text()

	if matches := reFindRule.FindStringSubmatch(line); matches != nil {
		// we found a rule so we need to advance the scanner to figure out if
		// there is a body
		beginLineNumber := scanner.LineNumber - 1
		scanner.Scan()
		bodyMatches := reFindRuleBody.FindStringSubmatch(scanner.Text())
		ruleBody := make([]string, 0, 20)
		for bodyMatches != nil {

			ruleBody = append(ruleBody, strings.TrimSpace(bodyMatches[1]))

			// done parsing the rule body line, advance the scanner and potentially
			// go into the next loop iteration
			scanner.Scan()
			bodyMatches = reFindRuleBody.FindStringSubmatch(scanner.Text())
		}
		// trim whitespace from all dependencies
		deps := strings.Split(matches[2], " ")
		filteredDeps := make([]string, 0, cap(deps))

		for idx := range deps {
			item := strings.TrimSpace(deps[idx])
			if item != "" {
				filteredDeps = append(filteredDeps, item)
			}
		}
		ret = Rule{
			Target:       strings.TrimSpace(matches[1]),
			Dependencies: filteredDeps,
			Body:         ruleBody,
			FileName:     scanner.FileHandle.Name(),
			LineNumber:   beginLineNumber,
		}
	} else if matches := reFindSimpleVariable.FindStringSubmatch(line); matches != nil {
		ret = Variable{
			Name:           strings.TrimSpace(matches[1]),
			Assignment:     strings.TrimSpace(matches[2]),
			SimplyExpanded: true,
			FileName:       scanner.FileHandle.Name(),
			LineNumber:     scanner.LineNumber,
		}
		scanner.Scan()
	} else if matches := reFindExpandedVariable.FindStringSubmatch(line); matches != nil {
		ret = Variable{
			Name:           strings.TrimSpace(matches[1]),
			Assignment:     strings.TrimSpace(matches[2]),
			SimplyExpanded: false,
			FileName:       scanner.FileHandle.Name(),
			LineNumber:     scanner.LineNumber,
		}
		scanner.Scan()
	} else {
		if strings.TrimSpace(line) != "" {
			logger.Debug(fmt.Sprintf("Unable to match line '%s' to a Rule or Variable", line))
		}
		scanner.Scan()
	}

	return
}
