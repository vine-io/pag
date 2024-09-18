package pag

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Prometheus *PrometheusConfig `json:"prometheus"`

	AlertManager *AlertManagerConfig `json:"alertManager"`

	Grafana *GrafanaConfig `json:"grafana"`
}

func (cfg *Config) Validate() error {
	if err := cfg.Prometheus.Validate(); err != nil {
		return fmt.Errorf("prometheus config: %w", err)
	}

	if err := cfg.AlertManager.Validate(); err != nil {
		return fmt.Errorf("alertManager config: %w", err)
	}

	if err := cfg.Grafana.Validate(); err != nil {
		return fmt.Errorf("grafana config: %w", err)
	}

	return nil
}

type PrometheusConfig struct {
	Host string `json:"host"`

	ConfigYAML string `json:"config_yaml"`
}

func (cfg *PrometheusConfig) Validate() error {
	if cfg.Host == "" {
		return errors.New("host is required")
	}

	if cfg.ConfigYAML == "" {
		return errors.New("config_yml is required")
	}

	if _, err := os.Stat(cfg.ConfigYAML); err != nil {
		return err
	}

	return nil
}

type AlertManagerConfig struct {
	Host string `json:"host"`

	ConfigYAML string `json:"config_yaml"`
}

func (cfg *AlertManagerConfig) Validate() error {
	if cfg.Host == "" {
		return errors.New("host is required")
	}

	if cfg.ConfigYAML == "" {
		return errors.New("config_yaml is required")
	}

	if _, err := os.Stat(cfg.ConfigYAML); err != nil {
		return err
	}

	return nil
}

type GrafanaConfig struct {
	Host string `json:"host"`

	APIToken string `json:"api_token"`
}

func (cfg *GrafanaConfig) Validate() error {
	if cfg.Host == "" {
		return errors.New("host is required")
	}

	if cfg.APIToken == "" {
		return errors.New("api_token is required")
	}

	return nil
}
