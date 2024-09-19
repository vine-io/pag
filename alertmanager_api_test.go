package pag

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getAlertmanagerAPI(t *testing.T) AlertManagerAPI {
	endpoint := os.Getenv("ALERTMANAGER_ENDPOINT")
	hc := &http.Client{}

	api, err := NewAlertManagerAPI(hc, &AlertManagerConfig{
		Endpoint:   endpoint,
		ConfigYAML: "testdata/alertmanager.yaml",
	})

	assert.NoError(t, err)
	return api
}

func TestNewAlertManagerAPI(t *testing.T) {
	api := getAlertmanagerAPI(t)

	cfg := api.ConfigYAML()

	assert.Equal(t, cfg.Receivers[0].WebhookConfigs[0].URL, "http://127.0.0.1:8000/webhook")
}
