package uniquetargets

import (
	"testing"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
	"github.com/stretchr/testify/assert"
)

func TestUniqueTargets(t *testing.T) {
	t.Parallel()
	makefile := parser.Makefile{
		FileName: "no-duplicates.mk",
		Rules: []parser.Rule{
			{Target: "all", LineNumber: 1},
			{Target: "test", LineNumber: 5},
			{Target: "clean", LineNumber: 9},
		},
	}

	rule := UniqueTargets{}
	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Equal(t, 0, len(ret), "no duplicates should produce no violations")
}

func TestNoUniqueTargetsDetected(t *testing.T) {
	t.Parallel()
	makefile := parser.Makefile{
		FileName: "unique_targets.mk",
		Rules: []parser.Rule{
			{Target: "test", LineNumber: 2},
			{Target: "build", LineNumber: 5},
			{Target: "test", LineNumber: 8}, // duplicate
		},
	}

	rule := UniqueTargets{}
	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Equal(t, 1, len(ret), "expected one duplicate violation")
	assert.Contains(t, ret[0].Violation, `"test" defined multiple times`)
	assert.Equal(t, "unique_targets.mk", ret[0].FileName)
	assert.Equal(t, 8, ret[0].LineNumber)
}

func TestUniqueTargetsIgnoredByConfig(t *testing.T) {
	t.Parallel()
	makefile := parser.Makefile{
		FileName: "unique_ignored.mk",
		Rules: []parser.Rule{
			{Target: "all", LineNumber: 1},
			{Target: "test", LineNumber: 3},
			{Target: "test", LineNumber: 6}, // duplicate but ignored
			{Target: "deploy", LineNumber: 9},
			{Target: "deploy", LineNumber: 12}, // not ignored
		},
	}

	rule := UniqueTargets{}
	cfg := rules.RuleConfig{"ignore": "test,clean"}

	ret := rule.Run(makefile, cfg)

	assert.Equal(t, 1, len(ret), "only non-ignored duplicates should trigger violations")
	assert.Contains(t, ret[0].Violation, `"deploy" defined multiple times`)
	assert.NotContains(t, ret[0].Violation, `"test"`)
}

func TestPhonyTargetsAreIgnored(t *testing.T) {
	t.Parallel()
	makefile := parser.Makefile{
		FileName: "phony_targets.mk",
		Rules: []parser.Rule{
			{Target: ".PHONY", LineNumber: 1},
			{Target: ".PHONY", LineNumber: 3},
			{Target: "build", LineNumber: 5},
			{Target: "build", LineNumber: 8}, // duplicate real target
		},
	}

	rule := UniqueTargets{}
	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Equal(t, 1, len(ret), "only non-.PHONY duplicates should trigger violations")
	assert.Contains(t, ret[0].Violation, `"build" defined multiple times`)
}
