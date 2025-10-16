// Package config deals with loading and parsing configuration from disk
package config

import (
	"fmt"

	"github.com/checkmake/checkmake/logger"
	"github.com/checkmake/checkmake/rules"
	"github.com/go-ini/ini"
)

// Config is a struct to configure the validator and rules
type Config struct {
	iniFile *ini.File
}

// NewConfigFromFile returns a config struct that is filled with the values
// from the passed in ini file
func NewConfigFromFile(path string) (*Config, error) {
	iniFile, err := ini.Load(path)
	ret := &Config{
		iniFile: iniFile,
	}

	return ret, err
}

// GetRuleConfig returns a rules.RuleConfig for the given rule. A rule
// corresponds to a section in the config ini file
func (c *Config) GetRuleConfig(rule string) (ret rules.RuleConfig) {

	if c.iniFile == nil {
		logger.Debug("iniFile not initialized")
		return
	}

	ret = make(rules.RuleConfig)

	section, err := c.iniFile.GetSection(rule)

	if err == nil {
		for _, keyName := range section.KeyStrings() {

			key, keyError := section.GetKey(keyName)
			if keyError == nil {
				ret[keyName] = key.String()
			}
		}
	}

	return
}

// GetConfigValue returns a configuration value from the config file from the
// default section. The way the configuration structure works for now is that
// sections are mostly for rules and values for checkmake itself are in the
// default section. That's why we just return values from the default section
// here.
func (c *Config) GetConfigValue(keyName string) (value string, err error) {
	if c.iniFile == nil {
		logger.Debug("iniFile not initialized")
		return "", fmt.Errorf("No config file open")
	}

	section, err := c.iniFile.GetSection("default")

	if err == nil {
		key, keyError := section.GetKey(keyName)
		if keyError != nil {
			return "", fmt.Errorf("key '%s' doesn't exist in config", keyName)
		}
		return key.String(), nil
	}

	return "", fmt.Errorf("config has no default section")
}

// Ini returns the underlying ini.File instance for debugging or advanced inspection.
func (c *Config) Ini() *ini.File {
	return c.iniFile
}
