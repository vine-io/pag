package pag

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getGrafanaAPI(t *testing.T) GrafanaAPI {
	endpoint := os.Getenv("GRAFANA_ENDPOINT")
	apiToken := os.Getenv("GRAFANA_API_TOKEN")
	username := os.Getenv("GRAFANA_USERNAME")
	password := os.Getenv("GRAFANA_PASSWORD")
	hc := &http.Client{}

	api, err := NewGrafanaAPI(hc, &GrafanaConfig{
		Endpoint: endpoint,
		APIToken: apiToken,
		Username: username,
		Password: password,
	})

	assert.NoError(t, err)
	return api
}

func TestNewGrafanaAPI(t *testing.T) {
	getGrafanaAPI(t)
}

func TestGrafanaAPI_FindOrCreateFolder(t *testing.T) {
	api := getGrafanaAPI(t)

	ctx := context.Background()
	folder, err := api.FindOrCreateFolder(ctx, "测试folder")
	if !assert.NoError(t, err) {
		return
	}

	out, err := api.DeleteFolder(ctx, folder.UID)
	if !assert.NoError(t, err) {
		return
	}

	t.Logf("delete folder result: %v", out)
}

func TestGrafanaAPI_GetDataSourceByName(t *testing.T) {
	api := getGrafanaAPI(t)

	ctx := context.Background()
	ds, err := api.GetDataSourceByName(ctx, "Prometheus")
	if !assert.NoError(t, err) {
		return
	}

	t.Logf("get data source result: %v", ds.Name)
}

func TestGetDashboardByUID(t *testing.T) {
	api := getGrafanaAPI(t)
	ctx := context.Background()

	dash, err := api.GetDashboardByUID(ctx, "test-node-dashboard-alerts")
	if !assert.NoError(t, err) {
		return
	}

	t.Logf("get dashboard result: %v", dash.Dashboard)
}
