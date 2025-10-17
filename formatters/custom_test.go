package formatters

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/checkmake/checkmake/config"
	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/validator"
	"github.com/stretchr/testify/assert"
)

func TestCustomFormatter(t *testing.T) {
	out := new(bytes.Buffer)

	tmpl, _ := template.New("test").Parse("{{.FileName}}:{{.LineNumber}}:{{.Rule}}:{{.Violation}}")
	formatter := CustomFormatter{template: tmpl, out: out}

	makefile, _ := parser.Parse("../fixtures/missing_phony.make")

	violations := validator.Validate(makefile, &config.Config{})
	formatter.Format(violations)
	assert.Regexp(t, `../fixtures/missing_phony.make:21:minphony:Required target "all" must be declared PHONY.`, out.String())
	assert.Regexp(t, `../fixtures/missing_phony.make:21:minphony:Required target "test" must be declared PHONY.`, out.String())
	assert.Regexp(t, `../fixtures/missing_phony.make:16:phonydeclared:Target "all" should be declared PHONY.`, out.String())
	assert.Equal(t, strings.Count(out.String(), "\n"), 3)
}

func TestCustomFormatterNewMethod(t *testing.T) {
	_, err := NewCustomFormatter("{{.FileName}}:{{.LineNumber}}:{{.Rule}}:{{.Violation}}")

	assert.Equal(t, nil, err)
}

func TestCustomFormatterNewMethodFailing(t *testing.T) {
	_, err := NewCustomFormatter("{{.LineNumber}}:{{.Rule}}:{{.Violation}}{{end}}")

	assert.NotEqual(t, nil, err)
}
