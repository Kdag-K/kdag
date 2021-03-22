package commands

import (
	conf "github.com/Kdag-K/kdag/src/config"
)

// CLIConfig contains configuration for the Run command
type CLIConfig struct {
	Kdag       conf.Config `mapstructure:",squash"`
	ProxyAddr  string      `mapstructure:"proxy-listen"`
	ClientAddr string      `mapstructure:"client-connect"`
}

// NewDefaultCLIConf creates a CLIConfig with default values
func NewDefaultCLIConf() *CLIConfig {
	return &CLIConfig{
		Kdag:       *conf.NewDefaultConf(),
		ProxyAddr:  "127.0.0.1:1338",
		ClientAddr: "127.0.0.1:1339",
	}
}
