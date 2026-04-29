package maxbodylength

import (
	"testing"

	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
	"github.com/stretchr/testify/assert"
)

func TestFooIsTooLong(t *testing.T) {
	makefile := parser.Makefile{
		FileName: "maxbodylength.mk",
		Rules: []parser.Rule{{
			Target: "foo",
			Body: []string{
				"echo 'foo'",
				"echo 'foo'",
				"echo 'foo'",
				"echo 'foo'",
				"echo 'foo'",
				"echo 'foo'",
				"echo 'foo'",
			},
			LineNumber: 1,
		}},
	}

	rule := MaxBodyLength{}

	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Equal(t, 1, len(ret))
	assert.Equal(t, "Target bodies should be kept simple and short (no more than 5 lines).",
		rule.Description(nil))
	assert.Equal(t, "Target body for \"foo\" exceeds allowed length of 5 lines (7).", ret[0].Violation)
	assert.Equal(t, 1, ret[0].LineNumber)
	assert.Equal(t, "maxbodylength.mk", ret[0].FileName)
}

func TestFooIsTooLongWithConfig(t *testing.T) {
	makefile := parser.Makefile{
		FileName: "maxbodylength.mk",
		Rules: []parser.Rule{{
			Target: "foo",
			Body: []string{
				"echo 'foo'",
				"echo 'foo'",
				"echo 'foo'",
				"echo 'foo'",
			},
			LineNumber: 1,
		}},
	}

	rule := MaxBodyLength{}

	cfg := rules.RuleConfig{}
	cfg["maxBodyLength"] = "3"

	ret := rule.Run(makefile, cfg)

	assert.Equal(t, 1, len(ret))
	assert.Equal(t, "Target bodies should be kept simple and short (no more than 3 lines).",
		rule.Description(nil))
	assert.Equal(t, "Target body for \"foo\" exceeds allowed length of 3 lines (4).", ret[0].Violation)
	assert.Equal(t, 1, ret[0].LineNumber)
	assert.Equal(t, "maxbodylength.mk", ret[0].FileName)
}

// TestMaxBodyLength_BackslashContinuation documents the secondary maxbodylength
// symptom of issue #244.
//
// Once the parser is fixed and correctly collects all three recipe lines
// (the tab-indented start line + two space-indented continuation lines) into
// the celerybeat body, maxbodylength must NOT fire — three lines is well within
// the default limit of 5.
//
// This test operates directly on a pre-built Makefile struct so that the
// rule logic can be verified independently of the parser fix.
func TestMaxBodyLength_BackslashContinuation(t *testing.T) {
	makefile := parser.Makefile{
		FileName: "backslash.mk",
		Rules: []parser.Rule{{
			Target: "celerybeat",
			// Three body lines: the initial tab-indented command plus the two
			// space-indented continuation lines — as the parser should return
			// after the issue #244 fix.
			Body: []string{
				`pipenv run celery -A myproject beat \`,
				`--scheduler django_celery_beat.schedulers:DatabaseScheduler \`,
				`--loglevel=INFO`,
			},
			LineNumber: 1,
		}},
	}

	rule := MaxBodyLength{}
	ret := rule.Run(makefile, rules.RuleConfig{})

	assert.Empty(t, ret,
		"three body lines is within the default limit of 5; no maxbodylength violation expected")
}
