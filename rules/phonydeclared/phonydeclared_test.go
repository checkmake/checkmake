package phonydeclared

import (
	"testing"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestPhonyDeclared_TargetSpecificVariables(t *testing.T) {
	t.Parallel()
	makefile, err := parser.Parse("../../fixtures/target_specific_variables.make")
	require.NoError(t, err)

	rule := Phonydeclared{}
	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Empty(t, ret, "target-specific variable assignments must not be reported as missing PHONY declarations")
}
