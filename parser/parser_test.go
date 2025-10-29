package parser

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/checkmake/checkmake/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSimpleMakefile(t *testing.T) {
	t.Parallel()
	ret, err := Parse("../fixtures/simple.make")

	assert.Equal(t, err, nil)
	assert.Equal(t, ret.FileName, "../fixtures/simple.make")
	assert.Equal(t, len(ret.Rules), 7)
	assert.Equal(t, len(ret.Variables), 2)
	assert.Equal(t, ret.Rules[0].Target, "clean")
	assert.Equal(t, ret.Rules[0].Body, []string{"rm bar", "rm foo"})
	assert.Equal(t, ret.Rules[0].FileName, "../fixtures/simple.make")

	assert.Equal(t, ret.Rules[1].Target, "foo")
	assert.Equal(t, ret.Rules[1].Body, []string{"touch foo"})
	assert.Equal(t, ret.Rules[1].Dependencies, []string{"bar"})
	assert.Equal(t, ret.Rules[1].FileName, "../fixtures/simple.make")

	assert.Equal(t, ret.Rules[2].Target, "bar")
	assert.Equal(t, ret.Rules[2].Body, []string{"touch bar"})
	assert.Equal(t, ret.Rules[2].FileName, "../fixtures/simple.make")

	assert.Equal(t, ret.Rules[3].Target, "all")
	assert.Equal(t, ret.Rules[3].Dependencies, []string{"foo"})
	assert.Equal(t, ret.Rules[3].FileName, "../fixtures/simple.make")

	assert.Equal(t, ".PHONY", ret.Rules[5].Target)
	assert.Equal(t, []string{"all", "clean", "test"}, ret.Rules[5].Dependencies)
	assert.Equal(t, "../fixtures/simple.make", ret.Rules[5].FileName)

	assert.Equal(t, ".DEFAULT_GOAL", ret.Rules[6].Target)
	assert.Equal(t, []string{"all"}, ret.Rules[6].Dependencies)
	assert.Equal(t, "../fixtures/simple.make", ret.Rules[6].FileName)

	assert.Equal(t, ret.Variables[0].Name, "expanded")
	assert.Equal(t, ret.Variables[0].Assignment, "\"$(simple)\"")
	assert.Equal(t, ret.Variables[0].SimplyExpanded, false)
	assert.Equal(t, ret.Variables[0].SpecialVariable, false)
	assert.Equal(t, ret.Variables[0].FileName, "../fixtures/simple.make")

	assert.Equal(t, ret.Variables[1].Name, "simple")
	assert.Equal(t, ret.Variables[1].Assignment, "\"foo\"")
	assert.Equal(t, ret.Variables[1].SimplyExpanded, true)
	assert.Equal(t, ret.Variables[1].SpecialVariable, false)
	assert.Equal(t, ret.Variables[1].FileName, "../fixtures/simple.make")
}

func TestParse_IgnoresEmptyLinesInDebug(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger.SetLogLevel(logger.DebugLevel)
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	_, err := Parse("../fixtures/unknown_lines.make")
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, "Unable to match line 'thisisnotarule'",
		"non-empty invalid lines should trigger a debug message")

	assert.NotContains(t, output, "Unable to match line ''",
		"empty lines should not trigger debug messages")
}

func writeTempMakefile(t *testing.T, content string) string {
	f, err := os.CreateTemp("", "*.mk")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func TestParse_VariableVsRuleMisclassification(t *testing.T) {
	t.Parallel()
	makefile := `
ifeq ($(OS), Windows_NT)
MKDIR := $(shell which mkdir.exe)
else
MKDIR := mkdir
endif
`
	tmp := writeTempMakefile(t, makefile)
	defer os.Remove(tmp)

	ret, err := Parse(tmp)
	require.NoError(t, err)

	varNames := []string{}
	for _, v := range ret.Variables {
		varNames = append(varNames, v.Name)
	}

	assert.Contains(t, varNames, "MKDIR", "MKDIR should be parsed as a variable, not a rule")
	assert.Empty(t, ret.Rules, "no rules should be parsed in this makefile")
}

func TestParse_BuildRuleRecognized(t *testing.T) {
	t.Parallel()
	makefile := `
build: clean ${BUILD_DIR}
	@echo building...
`
	tmp := writeTempMakefile(t, makefile)
	defer os.Remove(tmp)

	ret, err := Parse(tmp)
	require.NoError(t, err)

	assert.Len(t, ret.Rules, 1)
	assert.Equal(t, "build", ret.Rules[0].Target)
	assert.Contains(t, ret.Rules[0].Dependencies, "${BUILD_DIR}")
	assert.Equal(t, "@echo building...", ret.Rules[0].Body[0])
}

func TestParse_PatternAndPhonyRules(t *testing.T) {
	t.Parallel()
	makefile := `
%.o: %.c
	@echo compiling
.PHONY: all clean
`
	tmp := writeTempMakefile(t, makefile)
	defer os.Remove(tmp)

	ret, err := Parse(tmp)
	require.NoError(t, err)

	targets := []string{}
	for _, r := range ret.Rules {
		targets = append(targets, r.Target)
	}

	assert.Contains(t, targets, "%.o")
	assert.Contains(t, targets, ".PHONY")
}

func TestParse_TargetsWithSpecialChars(t *testing.T) {
	t.Parallel()
	makefile := `
target_with_underscores:
	echo "underscore"
target-with-hyphens  :
	echo "hyphen"
`
	tmp := writeTempMakefile(t, makefile)
	defer os.Remove(tmp)

	ret, err := Parse(tmp)
	require.NoError(t, err)

	targets := []string{}
	for _, r := range ret.Rules {
		targets = append(targets, r.Target)
	}

	assert.Contains(t, targets, "target_with_underscores")
	assert.Contains(t, targets, "target-with-hyphens")
}

func TestParse_MultipleTargetsAndSpaces(t *testing.T) {
	t.Parallel()
	makefile := `
target1 target2 : dep
	echo "build both"
`
	tmp := writeTempMakefile(t, makefile)
	defer os.Remove(tmp)

	ret, err := Parse(tmp)
	require.NoError(t, err)
	require.Len(t, ret.Rules, 1)

	assert.Equal(t, "target1 target2", ret.Rules[0].Target)
	assert.Contains(t, ret.Rules[0].Dependencies, "dep")
}

func TestParse_InlineRecipeAfterColon(t *testing.T) {
	t.Parallel()
	makefile := `
inline: dep ; echo "one-line recipe"
no_deps:; echo "no deps"
`
	tmp := writeTempMakefile(t, makefile)
	defer os.Remove(tmp)

	ret, err := Parse(tmp)
	require.NoError(t, err)

	assert.Equal(t, "inline", ret.Rules[0].Target)
	assert.Equal(t, []string{"dep"}, ret.Rules[0].Dependencies)
	assert.Contains(t, ret.Rules[0].Body[0], "echo")

	assert.Equal(t, "no_deps", ret.Rules[1].Target)
	assert.Empty(t, ret.Rules[1].Dependencies)
}

func TestParse_MultiTargetWithDepsAndInlineRecipe(t *testing.T) {
	t.Parallel()
	makefile := `
target1 target2 : dep1 dep2 ; echo "combo build"
	@echo "more lines"
`
	tmp := writeTempMakefile(t, makefile)
	defer os.Remove(tmp)

	ret, err := Parse(tmp)
	require.NoError(t, err)
	require.Len(t, ret.Rules, 1)

	rule := ret.Rules[0]
	assert.Equal(t, "target1 target2", rule.Target)
	assert.Equal(t, []string{"dep1", "dep2"}, rule.Dependencies)
	assert.Contains(t, rule.Body[0], "combo build")
	assert.Contains(t, rule.Body[1], "more lines")
}

func TestParse_FileExtAndVariableTargets(t *testing.T) {
	t.Parallel()
	makefile := `
file.ext:
	touch $@
$(DIR_VAR):
	echo "dir var"
${DIR_VAR}/subdir:
	echo "subdir rule"
`
	tmp := writeTempMakefile(t, makefile)
	defer os.Remove(tmp)

	ret, err := Parse(tmp)
	require.NoError(t, err)

	targets := []string{}
	for _, r := range ret.Rules {
		targets = append(targets, r.Target)
	}

	assert.Contains(t, targets, "file.ext")
	assert.Contains(t, targets, "$(DIR_VAR)")
	assert.Contains(t, targets, "${DIR_VAR}/subdir")
}

func TestParse_RuleWithEqualsInPrereq(t *testing.T) {
	t.Parallel()
	makefile := `
target: prerequisite = value
	@echo "rule with equals"
`
	tmp := writeTempMakefile(t, makefile)
	defer os.Remove(tmp)

	ret, err := Parse(tmp)
	require.NoError(t, err)

	require.Len(t, ret.Rules, 1)
	assert.Equal(t, "target", ret.Rules[0].Target)
	assert.Equal(t, []string{"prerequisite", "=", "value"}, ret.Rules[0].Dependencies)

	require.Len(t, ret.Rules[0].Body, 1)
	assert.Equal(t, "@echo \"rule with equals\"", ret.Rules[0].Body[0])
}

func TestParse_OtherVariableAssignments(t *testing.T) {
	t.Parallel()
	makefile := `
CONDITIONAL ?= default-value
SHELL_VAR != shell command
APPEND_VAR += more stuff
`
	tmp := writeTempMakefile(t, makefile)
	defer os.Remove(tmp)

	ret, err := Parse(tmp)
	require.NoError(t, err)

	vars := make(map[string]Variable)
	for _, v := range ret.Variables {
		vars[v.Name] = v
	}

	condVar, ok := vars["CONDITIONAL"]
	require.True(t, ok)
	assert.Equal(t, "default-value", condVar.Assignment)
	assert.False(t, condVar.SimplyExpanded, "?= should be recursive (false)")

	shellVar, ok := vars["SHELL_VAR"]
	require.True(t, ok)
	assert.Equal(t, "shell command", shellVar.Assignment)
	assert.True(t, shellVar.SimplyExpanded, "!= should be simple (true)")

	appendVar, ok := vars["APPEND_VAR"]
	require.True(t, ok)
	assert.Equal(t, "more stuff", appendVar.Assignment)
	assert.False(t, appendVar.SimplyExpanded, "+= should default to recursive (false)")
}

func TestParse_VariableLikeRuleIsNotRule(t *testing.T) {
	t.Parallel()
	makefile := `
EXTENSION := .exe
VAR ?= foo
APPEND += bar
SHELL != echo hi
`
	tmp := writeTempMakefile(t, makefile)
	defer os.Remove(tmp)

	ret, err := Parse(tmp)
	require.NoError(t, err)

	assert.Empty(t, ret.Rules, "variable assignments with ':=' or similar should not create rules")

	varNames := []string{}
	for _, v := range ret.Variables {
		varNames = append(varNames, v.Name)
	}
	assert.Contains(t, varNames, "EXTENSION")
	assert.Contains(t, varNames, "VAR")
	assert.Contains(t, varNames, "APPEND")
	assert.Contains(t, varNames, "SHELL")
}
