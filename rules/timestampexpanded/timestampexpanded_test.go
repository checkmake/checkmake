package timestampexpanded

import (
	"testing"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
	"github.com/stretchr/testify/assert"
)

func TestVersionIsNotSimplyExpanded(t *testing.T) {
	makefile := parser.Makefile{
		FileName: "timestamp-expanded.mk",
		Variables: []parser.Variable{{
			Name:           "BUILDTIME",
			Assignment:     "$(shell date -u +\"%Y-%m-%dT%H:%M:%SZ\")",
			SimplyExpanded: false,
		}},
	}

	rule := Timestampexpanded{}

	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Equal(t, 1, len(ret))
	assert.Equal(t, "timestamp variables should be simply expanded",
		rule.Description(nil))
	for i := range ret {
		assert.Equal(t, "timestamp-expanded.mk", ret[i].FileName)
	}
}

func TestVersionIsSimplyExpanded(t *testing.T) {
	makefile := parser.Makefile{
		FileName: "timestamp-simply-expanded.mk",
		Variables: []parser.Variable{{
			Name:           "BUILDTIME",
			Assignment:     "$(shell date -u +\"%Y-%m-%dT%H:%M:%SZ\")",
			SimplyExpanded: true,
		}},
	}

	rule := Timestampexpanded{}

	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Equal(t, 0, len(ret))
}
