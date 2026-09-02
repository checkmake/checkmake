package phonydeclared

import (
	"testing"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
	"github.com/stretchr/testify/assert"
)

func TestAllTargetsArePhony(t *testing.T) {
	t.Parallel()
	makefile := parser.Makefile{
		FileName: "phony-declared-all-phony.mk",
		Variables: []parser.Variable{{
			Name:       "PHONY",
			Assignment: "all clean",
		}},
		Rules: []parser.Rule{
			{
				Target: "all",
			}, {Target: "clean"},
		},
	}

	rule := Phonydeclared{}

	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Equal(t, len(ret), 0)
}

func TestMissingOnePhonyTarget(t *testing.T) {
	t.Parallel()
	makefile := parser.Makefile{
		FileName: "phony-declared-missing-one-phony.mk",
		Variables: []parser.Variable{{
			Name:       "PHONY",
			Assignment: "all",
		}},
		Rules: []parser.Rule{
			{
				Target: "all",
			}, {Target: "clean"},
		},
	}

	rule := Phonydeclared{}

	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Equal(t, len(ret), 1)

	for i := range ret {
		assert.Equal(t, "phony-declared-missing-one-phony.mk", ret[i].FileName)
	}
}

func TestPatternRuleSatisfiesBodylessTarget(t *testing.T) {
	t.Parallel()
	makefile := parser.Makefile{
		FileName: "phony-declared-pattern-rule.mk",
		Rules: []parser.Rule{
			// Extra-prerequisites-only rule: no body of its own, but
			// "%.o: %.c" below actually builds it. Must not be flagged.
			{Target: "foo.o", Dependencies: []string{"foo.h"}},
			{Target: "%.o", Dependencies: []string{"%.c"}, Body: []string{"$(CC) -c $< -o $@"}},
		},
	}

	rule := Phonydeclared{}

	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Equal(t, 0, len(ret))
}

func TestUnmatchedPatternRuleStillFlagged(t *testing.T) {
	t.Parallel()
	makefile := parser.Makefile{
		FileName: "phony-declared-pattern-rule-mismatch.mk",
		Rules: []parser.Rule{
			// "bar.txt" doesn't match "%.o", so it's still a real,
			// unexplained bodyless target: must be flagged.
			{Target: "bar.txt"},
			{Target: "%.o", Dependencies: []string{"%.c"}, Body: []string{"$(CC) -c $< -o $@"}},
		},
	}

	rule := Phonydeclared{}

	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Equal(t, 1, len(ret))
}

func TestExplicitRuleElsewhereSatisfiesBodylessTarget(t *testing.T) {
	t.Parallel()
	makefile := parser.Makefile{
		FileName: "phony-declared-explicit-elsewhere.mk",
		Rules: []parser.Rule{
			// Extra-prerequisites-only appearance of "foo.o", no body.
			{Target: "foo.o", Dependencies: []string{"foo.h"}},
			// The real recipe lives in a separate rule for the same target.
			{Target: "foo.o", Dependencies: []string{"foo.c"}, Body: []string{"$(CC) -c foo.c -o foo.o"}},
		},
	}

	rule := Phonydeclared{}

	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Equal(t, 0, len(ret))
}
