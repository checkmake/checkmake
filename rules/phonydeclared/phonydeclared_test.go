package phonydeclared

import (
	"os"
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

// TestPhonyDeclared_BackslashContinuation is the end-to-end regression test
// for issue #257 (https://github.com/checkmake/checkmake/issues/257).
//
// Space-indented backslash continuation lines in a recipe body must not be
// misidentified as rule targets by the parser, and phonydeclared must not
// fire a spurious violation for them.
func TestPhonyDeclared_BackslashContinuation(t *testing.T) {
	t.Parallel()

	content := ".PHONY: celerybeat\n" +
		"celerybeat: reset_redis\n" +
		"\tpipenv run celery -A myproject beat \\\n" +
		"        --scheduler django_celery_beat.schedulers:DatabaseScheduler \\\n" +
		"        --loglevel=INFO\n"

	f, err := os.CreateTemp("", "*.mk")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	defer os.Remove(f.Name())

	makefile, err := parser.Parse(f.Name())
	require.NoError(t, err)

	rule := Phonydeclared{}
	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Empty(t, ret,
		"no phonydeclared violations expected: continuation lines are not rule targets")
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
