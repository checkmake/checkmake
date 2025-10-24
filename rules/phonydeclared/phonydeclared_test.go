package phonydeclared

import (
	"testing"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
	"github.com/stretchr/testify/assert"
)

func TestAllTargetsArePhony(t *testing.T) {
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
