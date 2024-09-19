package pag

import (
	"net/http"
)

type Client struct {
	cfg *Config

	hc *http.Client
}

func NewClient(cfg *Config) (*Client, error) {
	return NewWithHttpClient(&http.Client{}, cfg)
}

func NewWithHttpClient(hc *http.Client, cfg *Config) (*Client, error) {
	c := &Client{
		hc:  hc,
		cfg: cfg,
	}

	return c, nil
}

func (c *Client) Prometheus() (PrometheusAPI, error) {
	return NewPrometheusAPI(c.hc, c.cfg.Prometheus)
}

func (c *Client) AlertManager() (AlertManagerAPI, error) {
	return NewAlertManagerAPI(c.hc, c.cfg.AlertManager)
}

func (c *Client) Grafana() (GrafanaAPI, error) {
	return NewGrafanaAPI(c.hc, c.cfg.Grafana)
}
