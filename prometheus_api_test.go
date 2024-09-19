package pag

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getPrometheusAPI(t *testing.T) PrometheusAPI {
	endpoint := os.Getenv("PROMETHEUS_ENDPOINT")
	hc := &http.Client{}

	api, err := NewPrometheusAPI(hc, &PrometheusConfig{
		Endpoint:   endpoint,
		ConfigYAML: "testdata/prometheus.yaml",
	})

	assert.NoError(t, err)
	return api
}

func TestNewPrometheusAPI(t *testing.T) {
	api := getPrometheusAPI(t)

	cfg := api.ConfigYAML()

	assert.Equal(t, cfg.ScrapeConfigs[0].JobName, "prometheus")
}

func TestPrometheusAPI_Values(t *testing.T) {
	api := getPrometheusAPI(t)

	ctx := context.Background()
	values, err := api.Values(ctx)

	assert.NoError(t, err)
	assert.NotEqual(t, len(values), 0)

	t.Logf("first value = %s", values[0])
}
