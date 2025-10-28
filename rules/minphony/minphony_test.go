package minphony

import (
	"fmt"
	"strings"
	"testing"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
	"github.com/stretchr/testify/assert"
)

var mpRunTests = []struct {
	mf parser.Makefile
	vl rules.RuleViolationList
}{
	{
		mf: parser.Makefile{
			FileName: "green-eggs.mk",
			Rules: parser.RuleList{
				{Target: "green-eggs"},
				{Target: "ham"},
			},
			Variables: parser.VariableList{
				{Name: "PHONY", Assignment: "green-eggs ham"},
			},
		},
		vl: rules.RuleViolationList{
			rules.RuleViolation{
				Rule:       "minphony",
				Violation:  "Required target \"kleen\" is missing from the Makefile.",
				FileName:   "green-eggs.mk",
				LineNumber: -1,
			},
			rules.RuleViolation{
				Rule:       "minphony",
				Violation:  "Required target \"awl\" is missing from the Makefile.",
				FileName:   "green-eggs.mk",
				LineNumber: -1,
			},
			rules.RuleViolation{
				Rule:       "minphony",
				Violation:  "Required target \"toast\" is missing from the Makefile.",
				FileName:   "green-eggs.mk",
				LineNumber: -1,
			},
		},
	},
	{
		mf: parser.Makefile{
			FileName: "kleen.mk",
			Rules: parser.RuleList{
				{Target: "awl"},
				{Target: "distkleen"},
				{Target: "kleen"},
			},
			Variables: parser.VariableList{
				{Name: "PHONY", Assignment: "awl kleen distkleen"},
			},
		},
		vl: rules.RuleViolationList{
			rules.RuleViolation{
				Rule:       "minphony",
				Violation:  "Required target \"toast\" is missing from the Makefile.",
				FileName:   "kleen.mk",
				LineNumber: -1,
			},
		},
	},
}

func TestMinPhony_new(t *testing.T) {
	t.Parallel()
	mp := &MinPhony{required: []string{"oh", "hai"}}

	assert.Equal(t, []string{"oh", "hai"}, mp.required)
	assert.Equal(t, "minphony", mp.Name())
	expectedDesc := fmt.Sprintf("Minimum required phony targets must be present (%s).", strings.Join(mp.required, ","))

	assert.Equal(t, expectedDesc, mp.Description(nil))
}

func TestMinPhony_Run(t *testing.T) {
	t.Parallel()
	mp := &MinPhony{required: []string{"kleen", "awl", "toast"}}

	for _, test := range mpRunTests {
		assert.Equal(t, test.vl, mp.Run(test.mf, rules.RuleConfig{}))
	}
}

func TestMinPhony_RunWithConfig(t *testing.T) {
	t.Parallel()
	mp := &MinPhony{required: []string{}}

	mf := parser.Makefile{
		FileName: "test.mk",
		Rules: parser.RuleList{
			{Target: "clone"},
			{Target: "toast"},
		},
		Variables: parser.VariableList{
			{Name: "PHONY", Assignment: "clone toast"},
		},
	}
	vl := rules.RuleViolationList{
		rules.RuleViolation{
			Rule:       "minphony",
			Violation:  "Required target \"foo\" is missing from the Makefile.",
			FileName:   "test.mk",
			LineNumber: -1,
		},
		rules.RuleViolation{
			Rule:       "minphony",
			Violation:  "Required target \"bar\" is missing from the Makefile.",
			FileName:   "test.mk",
			LineNumber: -1,
		},
	}
	cfg := rules.RuleConfig{}
	cfg["required"] = "foo, bar"

	assert.Equal(t, vl, mp.Run(mf, cfg))

	cfg["required"] = ""
	vl = rules.RuleViolationList{}

	assert.Equal(t, vl, mp.Run(mf, cfg))
}

func TestMinPhony_MissingPhonyDeclaration(t *testing.T) {
	t.Parallel()
	makefile := parser.Makefile{
		FileName: "missing-phony.mk",
		Rules: []parser.Rule{
			{Target: "all"},
			{Target: "clean"},
			{Target: "test"},
		},
		Variables: []parser.Variable{
			{Name: "PHONY", Assignment: "all"}, // only "all" declared
		},
	}

	mp := &MinPhony{required: []string{"all", "clean", "test"}}
	ret := mp.Run(makefile, rules.RuleConfig{})

	assert.Len(t, ret, 2, "expected two missing PHONY declaration violations")
	assert.Equal(t, "Required target \"clean\" must be declared PHONY.", ret[0].Violation)
	assert.Equal(t, "Required target \"test\" must be declared PHONY.", ret[1].Violation)
}
