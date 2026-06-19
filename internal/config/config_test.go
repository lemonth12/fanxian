package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	yamlContent := `
server:
  port: 8080
  mode: debug

database:
  path: ./data.db
  wal: true

jwt:
  access_expire: 2h
  refresh_expire: 168h
  secret: testsecret

jd_union:
  app_key: testkey
  app_secret: testsecret
  site_id: testsite
  pid: testpid

cashback:
  default_rate: 0.7

rate_limit:
  convert_per_minute: 10
`
	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(yamlContent)
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Cashback.DefaultRate != 0.7 {
		t.Errorf("rate = %f, want 0.7", cfg.Cashback.DefaultRate)
	}
	if cfg.JDUnion.AppKey != "testkey" {
		t.Errorf("app_key = %s, want testkey", cfg.JDUnion.AppKey)
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	os.Setenv("JWT_SECRET", "envsecret")
	defer os.Unsetenv("JWT_SECRET")

	yamlContent := `
jwt:
  secret: '${JWT_SECRET}'
jd_union:
  app_key: test
  app_secret: test
  site_id: test
  pid: test
`
	f, _ := os.CreateTemp("", "config-*.yaml")
	defer os.Remove(f.Name())
	f.WriteString(yamlContent)
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.JWT.Secret != "envsecret" {
		t.Errorf("secret = %s, want envsecret", cfg.JWT.Secret)
	}
}

func TestLoadConfig_MissingRequired(t *testing.T) {
	yamlContent := `
server:
  port: 8080
`
	f, _ := os.CreateTemp("", "config-*.yaml")
	defer os.Remove(f.Name())
	f.WriteString(yamlContent)
	f.Close()

	_, err := Load(f.Name())
	if err == nil {
		t.Error("expected error for missing jd_union config")
	}
}
