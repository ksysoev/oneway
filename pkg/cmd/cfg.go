package cmd

import (
	"fmt"
	"strings"

	"github.com/ksysoev/oneway/pkg/core/revconproxy"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Exchange *ExchaneConfig      `mapstructure:"exchange"`
	Revproxy *revconproxy.Config `mapstructure:"revproxy"`
}

func initConfig(configPath string) (*AppConfig, error) {
	viper.SetConfigFile(configPath)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &AppConfig{}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
