package config

import (
	"github.com/spf13/viper"
)

// Config 存储了应用的所有配置
type Config struct {
	CheckInterval int `mapstructure:"check_interval"`
	Systemd       struct {
		Services []string `mapstructure:"services"`
	} `mapstructure:"systemd"`
	Lark struct {
		WebhookURL string `mapstructure:"webhook_url"`
	} `mapstructure:"lark"`
}

// Load 从指定路径加载配置
func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv() // 允许通过环境变量覆盖配置

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}