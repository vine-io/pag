package pag

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/api"
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
	address := c.cfg.Prometheus.Host
	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}

	apiCfg := api.Config{
		Address: address,
		Client:  c.hc,
	}
	pc, err := api.NewClient(apiCfg)
	if err != nil {
		return nil, err
	}

	pa := &prometheusAPI{
		cfg: c.cfg.Prometheus,
		c:   pc,
	}

	if err = pa.load(); err != nil {
		return nil, err
	}

	return pa, nil
}

func (c *Client) AlertManager() (AlertManagerAPI, error) {
	pa := &alertManagerAPI{
		cfg: c.cfg.AlertManager,
		hc:  c.hc,
	}

	if err := pa.load(); err != nil {
		return nil, err
	}

	return pa, nil
}
