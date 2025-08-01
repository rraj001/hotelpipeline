package infra

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Hotels  HotelsConfig  `yaml:"hotels"`
	Redis   RedisConfig   `yaml:"redis"`
	CronJob CronJobConfig `yaml:"cronjob"`
	HTTP    HTTPConfig    `yaml:"http"`
}

type HotelsConfig struct {
	URLs []string `yaml:"urls"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	DB   int    `yaml:"db"`
}

type CronJobConfig struct {
	Interval string `yaml:"interval"`
}

type HTTPConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if len(c.Hotels.URLs) == 0 {
		return fmt.Errorf("at least one hotel supplier URL is required")
	}

	if c.Redis.Host == "" {
		return fmt.Errorf("Redis host is required")
	}

	if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
		return fmt.Errorf("Redis port must be between 1 and 65535")
	}

	if c.CronJob.Interval == "" {
		return fmt.Errorf("cron job interval is required")
	}

	if c.HTTP.Port <= 0 || c.HTTP.Port > 65535 {
		return fmt.Errorf("HTTP port must be between 1 and 65535")
	}

	if c.HTTP.Host == "" {
		return fmt.Errorf("HTTP host is required")
	}

	return nil
}
