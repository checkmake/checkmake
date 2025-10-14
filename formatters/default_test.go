package formatters

import (
	"bytes"
	"testing"

	"github.com/checkmake/checkmake/config"
	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/validator"
	"github.com/stretchr/testify/assert"
)

func TestDefaultFormatter(t *testing.T) {
	out := new(bytes.Buffer)
	formatter := DefaultFormatter{out: out}

	makefile, _ := parser.Parse("../fixtures/missing_phony.make")

	violations := validator.Validate(makefile, &config.Config{})
	formatter.Format(violations)

	assert.Regexp(t, `(?s)\s+RULE\s+DESCRIPTION\s+FILE NAME\s+LINE NUMBER\s+`, out.String())
	assert.Regexp(t, `(?s)phonydeclared\s+Target "all".+\s+16`, out.String())
	assert.Regexp(t, `(?s)declared\s+PHONY`, out.String())
}
