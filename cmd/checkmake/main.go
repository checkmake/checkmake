package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/checkmake/checkmake/config"
	"github.com/checkmake/checkmake/formatters"
	"github.com/checkmake/checkmake/logger"
	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
	"github.com/checkmake/checkmake/validator"
	docopt "github.com/docopt/docopt-go"
	"github.com/olekukonko/tablewriter"
)

var (
	usage = `checkmake.

  Usage:
  checkmake [options] [<makefile>...]
  checkmake -h | --help
  checkmake --version
  checkmake --list-rules

  Options:
  -h --help               Show this screen.
  --version               Show version.
  --debug                 Enable debug mode
  --config=<configPath>   Configuration file to read
  --format=<format>       Output format as a Golang text/template template
  --list-rules            List registered rules
`

	version   = ""
	buildTime = ""
	builder   = ""
	goversion = ""

	configPath = "checkmake.ini"
)

func main() {

	args, err := docopt.Parse(usage, nil, true,
		fmt.Sprintf("%s %s built at %s by %s with %s",
			"checkmake", version, buildTime, builder, goversion), false)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	formatter, violations := parseArgsAndGetFormatter(args)

	if len(violations) > 0 {
		formatter.Format(violations)
	}

	os.Exit(len(violations))
}

func parseArgsAndGetFormatter(args map[string]interface{}) (formatters.Formatter,
	rules.RuleViolationList) {
	if args["--debug"] == true {
		logger.SetLogLevel(logger.DebugLevel)
	}

	if args["--list-rules"] == true {
		listRules(os.Stdout)
		os.Exit(0)
	}

	if args["--config"] != nil {
		configPath = args["--config"].(string)
	} else {
		_, err := os.Stat(configPath)
		if os.IsNotExist(err) {
			home := os.Getenv("HOME")
			configPath = home + "/checkmake.ini"
		}
	}

	cfg, cfgError := config.NewConfigFromFile(configPath)

	if cfgError != nil {
		logger.Info(fmt.Sprintf("Unable to parse config file %q, running with defaults",
			configPath))
	}

	var violations rules.RuleViolationList
	makefileArray := args["<makefile>"].([]string)
	logger.Debug(fmt.Sprintf("Makefiles passed: %q",
		makefileArray))
	for _, mkf := range makefileArray {
		logger.Info(fmt.Sprintf("Parsing file %q",
			mkf))
		makefile, parseError := parser.Parse(mkf)
		if parseError != nil {
			log.Fatal(parseError)
			os.Exit(1)
		}
		violations = append(violations, validator.Validate(makefile, cfg)...)
	}

	var formatter formatters.Formatter

	if args["--format"] != nil {
		format := args["--format"].(string)
		var err error
		formatter, err = formatters.NewCustomFormatter(format)
		if err != nil {
			logger.Error(fmt.Sprintf("Unable to create formatter: %q", err.Error()))
			os.Exit(1)
		}
	} else if format, formatErr := cfg.GetConfigValue("format"); formatErr == nil {
		var err error
		formatter, err = formatters.NewCustomFormatter(format)
		if err != nil {
			logger.Error(fmt.Sprintf("Unable to create formatter: %q", err.Error()))
			os.Exit(1)
		}

	} else {
		formatter = formatters.NewDefaultFormatter()
	}

	return formatter, violations
}

func listRules(w io.Writer) {
	data := [][]string{}
	for _, rule := range rules.GetRulesSorted() {
		data = append(data, []string{rule.Name(), rule.Description()})
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Name", "Description"})
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetRowSeparator(" ")
	table.SetAutoWrapText(true)

	table.AppendBulk(data)
	table.Render()
}
