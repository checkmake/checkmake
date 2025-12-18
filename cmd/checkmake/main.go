package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/checkmake/checkmake/config"
	"github.com/checkmake/checkmake/formatters"
	"github.com/checkmake/checkmake/logger"
	"github.com/checkmake/checkmake/parser"
	"github.com/checkmake/checkmake/rules"
	"github.com/checkmake/checkmake/validator"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

var (
	version   = ""
	buildTime = ""
	builder   = ""
	goversion = ""

	cfgPath     string
	debug       bool
	format      string
	output      string
	includeInfo bool
	errorsOnly  bool
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "checkmake [flags] [makefile...]",
		Short:        "Validate Makefiles for common issues",
		Long:         "checkmake scans Makefiles and reports potential issues according to configurable rules.",
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_ = cmd.Help()
				return nil
			}
			return runCheckmake(args)
		},
	}

	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode")
	cmd.PersistentFlags().StringVar(&cfgPath, "config", "checkmake.ini", "Configuration file to read")
	cmd.PersistentFlags().StringVar(&format, "format", "", "Custom Go template for text output (ignored in JSON mode)")
	cmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "Output format: 'text' (default) or 'json' (mutually exclusive with --format)")
	cmd.MarkFlagsMutuallyExclusive("format", "output")
	cmd.PersistentFlags().BoolVar(&includeInfo, "include-info", false, "Include info-level violations (style suggestions)")
	cmd.PersistentFlags().BoolVar(&errorsOnly, "errors-only", false, "Show only error-level violations (suppress warnings and info)")
	cmd.MarkFlagsMutuallyExclusive("include-info", "errors-only")

	cmd.Version = fmt.Sprintf("%s %s built at %s by %s with %s",
		"checkmake", version, buildTime, builder, goversion)

	cmd.AddCommand(&cobra.Command{
		Use:   "list-rules",
		Short: "List registered rules",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadConfig()
			listRules(cmd.OutOrStdout(), cfg)
		},
	})

	return cmd
}

func loadConfig() *config.Config {
	if debug {
		logger.SetLogLevel(logger.DebugLevel)
	}

	if _, err := os.Stat(cfgPath); err != nil {
		if os.IsNotExist(err) {
			home := os.Getenv("HOME")
			cfgPath = filepath.Join(home, "checkmake.ini")
		} else {
			logger.Error(fmt.Sprintf("error accessing config file %q: %v", cfgPath, err))
			return &config.Config{}
		}
	}

	cfg, cfgError := config.NewConfigFromFile(cfgPath)
	if cfgError != nil {
		logger.Info(fmt.Sprintf("Unable to parse config file %q, running with defaults", cfgPath))
		return &config.Config{}
	}

	logger.Debug(fmt.Sprintf("Using configuration file: %q", cfgPath))
	if debug {
		if iniFile := cfg.Ini(); iniFile != nil {
			var buf bytes.Buffer
			if _, err := iniFile.WriteTo(&buf); err == nil {
				logger.Debug(fmt.Sprintf("Parsed configuration:\n%s", buf.String()))
			}
		}
	}

	return cfg
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func runCheckmake(makefiles []string) error {
	cfg := loadConfig()
	logger.Debug(fmt.Sprintf("Makefiles passed: %q", makefiles))

	var allViolations rules.RuleViolationList
	for _, mkf := range makefiles {
		logger.Info(fmt.Sprintf("Parsing file %q", mkf))
		makefile, parseErr := parser.Parse(mkf)
		if parseErr != nil {
			return fmt.Errorf("failed to parse %q: %w", mkf, parseErr)
		}
		allViolations = append(allViolations, validator.Validate(makefile, cfg)...)
	}

	// Filter violations based on CLI flags
	var violations rules.RuleViolationList
	var errorCount int
	for _, v := range allViolations {
		if errorsOnly && v.Severity != rules.SeverityError {
			continue
		}
		if !includeInfo && v.Severity == rules.SeverityInfo {
			continue
		}
		violations = append(violations, v)
		if v.Severity == rules.SeverityError {
			errorCount++
		}
	}

	var formatter formatters.Formatter
	var err error

	// Priority: format flag > output flag > config format > default
	if format != "" {
		formatter, err = formatters.NewCustomFormatter(format)
	} else {
		// Use output flag if specified, otherwise check config
		outputMode := strings.ToLower(output)
		if outputMode == "" {
			if o, oerr := cfg.GetConfigValue("output"); oerr == nil {
				outputMode = strings.ToLower(o)
			} else {
				outputMode = "text" // default
			}
		}

		switch outputMode {
		case "json":
			formatter = formatters.NewJSONFormatter()
		case "text":
			if format != "" {
				formatter, err = formatters.NewCustomFormatter(format)
			} else if f, ferr := cfg.GetConfigValue("format"); ferr == nil {
				formatter, err = formatters.NewCustomFormatter(f)
			} else {
				formatter = formatters.NewDefaultFormatter()
			}
		default:
			return fmt.Errorf("invalid output format: %q (supported: text, json)", outputMode)
		}
		logger.Debug(fmt.Sprintf("Using output mode: %s", outputMode))
	}
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to create formatter: %q", err.Error()))
		return err
	}

	// Output violations
	if len(violations) > 0 {
		formatter.Format(violations)
	}

	// Exit code: 1 only if there are errors, 0 otherwise
	if errorCount > 0 {
		return fmt.Errorf("errors found (%d)", errorCount)
	}

	return nil
}

func listRules(w io.Writer, cfg *config.Config) {
	data := [][]string{}
	for _, rule := range rules.GetRulesSorted() {
		cfgForRule := cfg.GetRuleConfig(rule.Name())
		// Get severity from config if set, otherwise use default
		severity := rules.GetSeverityFromConfig(rule, cfgForRule)
		data = append(data, []string{rule.Name(), string(severity), rule.Description(cfgForRule)})
	}

	table := tablewriter.NewTable(w,
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.BorderNone,
			Symbols: tw.NewSymbols(tw.StyleNone),
			Settings: tw.Settings{
				Lines:      tw.LinesNone,
				Separators: tw.SeparatorsNone,
			},
		}),
		tablewriter.WithRowAutoWrap(tw.WrapNormal),
		tablewriter.WithMaxWidth(72),
	)
	table.Header("Name", "Severity", "Description")

	if err := table.Bulk(data); err != nil {
		log.Fatalf("Bulk append failed: %v", err)
	}
	table.Render()
}
