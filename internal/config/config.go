package config

import (
	"github.com/spf13/viper"
)

type JournalConfig struct {
	ServiceName string   `yaml:"serviceName"`
	Keywords    []string `yaml:"keywords"`
}

type Config struct {
	CheckInterval int `yaml:"checkInterval"` // 修改为 yaml 标签，匹配配置文件
	Systemd       struct {
		Services []string `yaml:"services"`
	} `yaml:"systemd"`
	Lark struct {
		WebhookURLs []string `yaml:"webhook_urls"`
	} `yaml:"lark"`
	Journal []JournalConfig `yaml:"journal"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
