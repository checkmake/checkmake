package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

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

	cfgPath string
	debug   bool
	format  string
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
	cmd.PersistentFlags().StringVar(&format, "format", "", "Output format as a Go text/template template")

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
		logger.Error(fmt.Sprintf("Unable to parse config file %q, running with defaults", cfgPath))
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

	var violations rules.RuleViolationList
	for _, mkf := range makefiles {
		logger.Info(fmt.Sprintf("Parsing file %q", mkf))
		makefile, parseErr := parser.Parse(mkf)
		if parseErr != nil {
			return fmt.Errorf("failed to parse %q: %w", mkf, parseErr)
		}
		violations = append(violations, validator.Validate(makefile, cfg)...)
	}

	var formatter formatters.Formatter
	var err error

	if format != "" {
		formatter, err = formatters.NewCustomFormatter(format)
	} else if f, ferr := cfg.GetConfigValue("format"); ferr == nil {
		formatter, err = formatters.NewCustomFormatter(f)
	} else {
		formatter = formatters.NewDefaultFormatter()
	}
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to create formatter: %q", err.Error()))
		return err
	}

	// Output
	if len(violations) > 0 {
		formatter.Format(violations)
		return fmt.Errorf("violations found (%d)", len(violations))
	}

	return nil
}

func listRules(w io.Writer, cfg *config.Config) {
	data := [][]string{}
	for _, rule := range rules.GetRulesSorted() {
		cfgForRule := cfg.GetRuleConfig(rule.Name())
		data = append(data, []string{rule.Name(), rule.Description(cfgForRule)})
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
	table.Header("Name", "Description")

	if err := table.Bulk(data); err != nil {
		log.Fatalf("Bulk append failed: %v", err)
	}
	table.Render()
}
