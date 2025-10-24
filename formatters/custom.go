package formatters

import (
	"io"
	"os"
	"text/template"

	"github.com/checkmake/checkmake/logger"
	"github.com/checkmake/checkmake/rules"
)

// CustomFormatter is a formatter that is configurable via a template string
type CustomFormatter struct {
	out      io.Writer
	template *template.Template
}

// NewCustomFormatter returns a CustomFormatter struct
func NewCustomFormatter(templateString string) (ret *CustomFormatter, err error) {
	var tmpl *template.Template
	ret = &CustomFormatter{}
	tmpl, err = template.New("CustomFormatter").Parse(templateString)
	if err != nil {
		return ret, err
	}

	ret.template = tmpl
	ret.out = os.Stdout

	return
}

// Format is the function to call to get the formatted output
func (f *CustomFormatter) Format(violations rules.RuleViolationList) {
	for _, val := range violations {
		err := f.template.Execute(f.out, val)
		f.out.Write([]byte("\n"))
		if err != nil {
			logger.Error(err.Error())
		}
	}
}
