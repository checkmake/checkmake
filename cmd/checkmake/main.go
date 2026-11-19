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

	cfgPath string
	debug   bool
	format  string
	output  string
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
	cmd.PersistentFlags().StringVar(&cfgPath, "config", defaultConfigPath, "Configuration file to read")
	cmd.PersistentFlags().StringVar(&format, "format", "", "Custom Go template for text output (ignored in JSON mode)")
	cmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "Output format: 'text' (default) or 'json' (mutually exclusive with --format)")
	cmd.MarkFlagsMutuallyExclusive("format", "output")

	cmd.Version = fmt.Sprintf("%s built at %s by %s with %s",
		version, buildTime, builder, goversion)
	cmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")

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

// defaultConfigPath is the name of the config file that checkmake picks up
// when no --config flag is given.
const defaultConfigPath = "checkmake.ini"

// configCandidates returns the config files to try, in order of priority. The
// hidden variant of the default name is accepted next to the visible one, both
// in the current directory and in the home directory.
func configCandidates(path, home string) []string {
	candidates := []string{path}
	if path == defaultConfigPath {
		candidates = append(candidates, "."+defaultConfigPath)
	}
	if home != "" {
		candidates = append(candidates,
			filepath.Join(home, defaultConfigPath),
			filepath.Join(home, "."+defaultConfigPath),
		)
	}
	return candidates
}

// resolveConfigPath returns the first existing file out of configCandidates. If
// none of them exists it returns the global fallback in the home directory, so
// that the caller reports a missing config in the same way as before.
func resolveConfigPath(path, home string) (string, error) {
	for _, candidate := range configCandidates(path, home) {
		_, err := os.Stat(candidate)
		if err == nil {
			return candidate, nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("error accessing config file %q: %w", candidate, err)
		}
	}
	return filepath.Join(home, defaultConfigPath), nil
}

func loadConfig() *config.Config {
	if debug {
		logger.SetLogLevel(logger.DebugLevel)
	}

	resolved, err := resolveConfigPath(cfgPath, os.Getenv("HOME"))
	if err != nil {
		logger.Error(err.Error())
		return &config.Config{}
	}
	cfgPath = resolved

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

	// Output
	if len(violations) > 0 {
		formatter.Format(violations)
		return fmt.Errorf("violations found (%d)", len(violations))
	}

	return nil
}

func listRules(w io.Writer, cfg *config.Config) {
	rulesSorted := rules.GetRulesSorted()
	data := make([][]string, len(rulesSorted))

	for i, rule := range rulesSorted {
		cfgForRule := cfg.GetRuleConfig(rule.Name())
		data[i] = []string{
			rule.Name(),
			rule.Description(cfgForRule),
		}
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
