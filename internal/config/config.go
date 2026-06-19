package config

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
	WAL  bool   `yaml:"wal"`
}

type JWTConfig struct {
	AccessExpire  time.Duration `yaml:"access_expire"`
	RefreshExpire time.Duration `yaml:"refresh_expire"`
	Secret        string        `yaml:"secret"`
}

type JDUnionConfig struct {
	AppKey    string `yaml:"app_key"`
	AppSecret string `yaml:"app_secret"`
	SiteID    string `yaml:"site_id"`
	PID       string `yaml:"pid"`
}

type CashbackConfig struct {
	DefaultRate float64 `yaml:"default_rate"`
}

type RateLimitConfig struct {
	ConvertPerMinute int `yaml:"convert_per_minute"`
}

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	JWT       JWTConfig       `yaml:"jwt"`
	JDUnion   JDUnionConfig   `yaml:"jd_union"`
	Cashback  CashbackConfig  `yaml:"cashback"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

var envRe = regexp.MustCompile(`\${([^}]+)}`)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	content := envRe.ReplaceAllStringFunc(
		string(data),
		func(s string) string {
			key := envRe.FindStringSubmatch(s)[1]
			return os.Getenv(key)
		})

	cfg := &Config{}
	if err := yaml.Unmarshal([]byte(content), cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.JDUnion.AppKey == "" ||
		cfg.JDUnion.AppSecret == "" ||
		cfg.JDUnion.SiteID == "" ||
		cfg.JDUnion.PID == "" {
		return nil, fmt.Errorf("missing required jd_union config")
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Cashback.DefaultRate == 0 {
		cfg.Cashback.DefaultRate = 0.7
	}

	return cfg, nil
}
