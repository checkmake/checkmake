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

// TestPhonyDeclared_BackslashContinuation is the rule-level regression test
// for issue #244 (comment https://github.com/checkmake/checkmake/issues/244#issuecomment-4344766044).
//
// The parser currently misidentifies the space-indented continuation line
//
//   "        --scheduler django_celery_beat.schedulers:DatabaseScheduler \\"
//
// as a rule target because it is not consumed by the tab-only body-collection
// loop and its colon matches reFindRule.  phonydeclared then fires:
//
//   Target "--scheduler django_celery_beat.schedulers" should be declared PHONY.
//
// This test builds the Makefile struct that the parser *currently* produces
// (the buggy state) and asserts that phonydeclared must NOT raise a violation
// for the spurious target.  The test will FAIL until the parser is fixed.
func TestPhonyDeclared_BackslashContinuation(t *testing.T) {
	t.Parallel()

	// This is the struct the parser currently produces for:
	//
	//   .PHONY: celerybeat
	//   celerybeat: reset_redis
	//   \tpipenv run celery -A myproject beat \
	//           --scheduler django_celery_beat.schedulers:DatabaseScheduler \
	//           --loglevel=INFO
	//
	// BUG: the spurious rule whose target is
	// "--scheduler django_celery_beat.schedulers" should not exist.
	makefile := parser.Makefile{
		FileName: "celery.mk",
		Rules: []parser.Rule{
			{
				Target:       ".PHONY",
				Dependencies: []string{"celerybeat"},
			},
			{
				Target:       "celerybeat",
				Dependencies: []string{"reset_redis"},
				// BUG: currently only the first tab-indented line ends up here.
				Body: []string{`pipenv run celery -A myproject beat \`},
			},
			// BUG: these two rules should not exist – they are continuation
			// lines that were misidentified as targets by the parser.
			{
				Target: "--scheduler django_celery_beat.schedulers",
				Body:   []string{},
			},
		},
	}

	rule := Phonydeclared{}
	ret := rule.Run(makefile, rules.RuleConfig{})

	// After the fix the parser will not produce the spurious rule, so
	// phonydeclared will have nothing to complain about.
	// FAILING until issue #244 is fixed in the parser.
	for _, v := range ret {
		if v.Violation == `Target "--scheduler django_celery_beat.schedulers" should be declared PHONY.` {
			t.Errorf("BUG #244: spurious phonydeclared violation for continuation line fragment: %s", v.Violation)
		}
	}

	assert.Empty(t, ret,
		"no phonydeclared violations expected: the continuation line is not a real target")
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
