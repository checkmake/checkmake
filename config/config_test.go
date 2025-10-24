// Package config testing
package config

import (
	"testing"

	"github.com/checkmake/checkmake/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleConfig(t *testing.T) {
	cfg, err := NewConfigFromFile("../fixtures/exampleConfig.ini")

	require.Equal(t, nil, err, "Parsing of the fixture config file should have worked.")

	ruleCfg := cfg.GetRuleConfig("phonydeclared")

	assert.Equal(t, "true", ruleCfg["disabled"])
	assert.Equal(t, "bla", ruleCfg["foo"])
}

func TestFailConfig(t *testing.T) {
	cfg, err := NewConfigFromFile("../fixtures/idontexist.ini")

	assert.NotEqual(t, nil, err)

	val := cfg.GetRuleConfig("bla")

	assert.Equal(t, rules.RuleConfig(nil), val)
}

func TestGetConfigValue(t *testing.T) {
	cfg, err := NewConfigFromFile("../fixtures/exampleConfig.ini")

	require.Equal(t, nil, err, "Parsing of the fixture config file should have worked.")

	format, err := cfg.GetConfigValue("format")

	require.Equal(t, nil, err, "Getting a default format config value should have worked.")

	assert.Equal(t, "{{.LineNumber}}:{{.Rule}}:{{.Violation}}", format)
}

func TestGetConfigValueOnMissingConfigFile(t *testing.T) {
	cfg, err := NewConfigFromFile("../fixtures/idontexist.ini")

	assert.NotEqual(t, nil, err)

	format, err := cfg.GetConfigValue("format")

	assert.NotEqual(t, nil, err)

	assert.Equal(t, "", format)
	assert.Equal(t, "No config file open", err.Error())
}

func TestGetMissingConfigValue(t *testing.T) {
	cfg, err := NewConfigFromFile("../fixtures/exampleConfig.ini")

	require.Equal(t, nil, err, "Parsing of the fixture config file should have worked.")

	format, err := cfg.GetConfigValue("nothinghere")

	assert.NotEqual(t, nil, err)
	assert.Equal(t, "", format)
	assert.Equal(t, "key 'nothinghere' doesn't exist in config", err.Error())
}

func TestGetConfigValueOnMissingDefaultSection(t *testing.T) {
	cfg, err := NewConfigFromFile("../fixtures/exampleConfigNoDefault.ini")

	require.Equal(t, nil, err, "Parsing of the fixture config file should have worked.")

	format, err := cfg.GetConfigValue("format")

	assert.NotEqual(t, nil, err)

	assert.Equal(t, "", format)
	assert.Equal(t, "config has no default section", err.Error())
}
