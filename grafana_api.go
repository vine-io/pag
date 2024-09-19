package pag

import (
	"context"
	"net/http"
	urlpkg "net/url"
	"strings"

	"github.com/go-openapi/strfmt"
	goapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/api_keys"
	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/datasources"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/models"
)

type GrafanaAPI interface {
	AddAPIKey(ctx context.Context, name string, ttl int64) (string, error)

	GetDataSourceByName(ctx context.Context, name string) (*models.DataSource, error)
	AddDataSource(ctx context.Context, ds *models.DataSource) (*models.DataSource, error)

	FindOrCreateFolder(ctx context.Context, title string) (*models.Folder, error)
	DeleteFolder(ctx context.Context, uid string) (string, error)

	GetDashboardByUID(ctx context.Context, uid string) ([]byte, error)
	UpsertDashboard(ctx context.Context, folderUID string, dash any) ([]byte, error)
	DeleteDashboard(ctx context.Context, uid string) (string, error)
}

type grafanaAPI struct {
	cfg *GrafanaConfig
	hc  *http.Client

	gc *goapi.GrafanaHTTPAPI
}

func NewGrafanaAPI(hc *http.Client, cfg *GrafanaConfig) (GrafanaAPI, error) {
	address := cfg.Endpoint
	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}

	url, err := urlpkg.Parse(address)
	if err != nil {
		return nil, err
	}

	tc := &goapi.TransportConfig{
		// Host is the doman name or IP address of the host that serves the API.
		Host: url.Host,
		// BasePath is the URL prefix for all API paths, relative to the host root.
		BasePath: "/api",
		// Schemes are the transfer protocols used by the API (http or https).
		Schemes: []string{url.Scheme},
		// APIKey is an optional API key or service account token.
		APIKey: cfg.APIToken,
		// BasicAuth is optional basic auth credentials.
		BasicAuth: urlpkg.UserPassword(cfg.Username, cfg.Password),
		// OrgID provides an optional organization ID.
		// OrgID is only supported with BasicAuth since API keys are already org-scoped.
		OrgID: 1,
		//TLSConfig: &tls.Config{},
		// NumRetries contains the optional number of attempted retries
		NumRetries: 3,
		// RetryTimeout sets an optional time to wait before retrying a request
		RetryTimeout: 0,
		// RetryStatusCodes contains the optional list of status codes to retry
		// Use "x" as a wildcard for a single digit (default: [429, 5xx])
		RetryStatusCodes: []string{"420", "5xx"},
		// HTTPHeaders contains an optional map of HTTP headers to add to each request
		HTTPHeaders: map[string]string{},
	}

	gc := goapi.NewHTTPClientWithConfig(strfmt.Default, tc)
	_, err = gc.APIKeys.GetAPIkeys(&api_keys.GetAPIkeysParams{
		Context:    context.TODO(),
		HTTPClient: hc,
	})
	if err != nil {
		return nil, err
	}

	api := &grafanaAPI{
		cfg: cfg,
		hc:  hc,
		gc:  gc,
	}

	return api, nil
}

func (api *grafanaAPI) AddAPIKey(ctx context.Context, name string, ttl int64) (string, error) {
	params := &api_keys.AddAPIkeyParams{
		Body: &models.AddAPIKeyCommand{
			Name:          name,
			Role:          models.AddAPIKeyCommandRoleAdmin,
			SecondsToLive: ttl,
		},
		Context:    ctx,
		HTTPClient: api.hc,
	}

	rsp, err := api.gc.APIKeys.AddAPIkeyWithParams(params)
	if err != nil {
		return "", err
	}
	return rsp.Payload.Key, nil
}

func (api *grafanaAPI) GetDataSourceByName(ctx context.Context, name string) (*models.DataSource, error) {
	params := &datasources.GetDataSourceByNameParams{
		Name:       name,
		Context:    ctx,
		HTTPClient: api.hc,
	}

	rsp, err := api.gc.Datasources.GetDataSourceByNameWithParams(params)
	if err != nil {
		return nil, err
	}

	return rsp.Payload, nil
}

func (api *grafanaAPI) AddDataSource(ctx context.Context, ds *models.DataSource) (*models.DataSource, error) {
	params := &datasources.AddDataSourceParams{
		Body: &models.AddDataSourceCommand{
			Access:          ds.Access,
			BasicAuth:       ds.BasicAuth,
			BasicAuthUser:   ds.BasicAuthUser,
			Database:        ds.Database,
			IsDefault:       ds.IsDefault,
			JSONData:        ds.JSONData,
			Name:            ds.Name,
			Type:            ds.Type,
			UID:             ds.UID,
			URL:             ds.URL,
			User:            ds.User,
			WithCredentials: ds.WithCredentials,
		},
		Context:    ctx,
		HTTPClient: api.hc,
	}

	rsp, err := api.gc.Datasources.AddDataSourceWithParams(params)
	if err != nil {
		return nil, err
	}
	return rsp.Payload.Datasource, nil
}

func (api *grafanaAPI) FindOrCreateFolder(ctx context.Context, title string) (*models.Folder, error) {
	getFolderRsp, err := api.gc.Folders.GetFolders(&folders.GetFoldersParams{
		Context:    nil,
		HTTPClient: nil,
	})
	if err != nil {
		return nil, err
	}

	var uid string
	for _, item := range getFolderRsp.Payload {
		if item.Title == title {
			uid = item.UID
		}
	}

	if uid != "" {
		rsp, err := api.gc.Folders.GetFolderByUIDWithParams(&folders.GetFolderByUIDParams{
			FolderUID:  uid,
			Context:    ctx,
			HTTPClient: api.hc,
		})
		if err != nil {
			return nil, err
		}
		return rsp.Payload, nil
	}

	rsp, err := api.gc.Folders.CreateFolderWithParams(&folders.CreateFolderParams{
		Body: &models.CreateFolderCommand{
			Title: title,
		},
		Context:    ctx,
		HTTPClient: api.hc,
	})
	if err != nil {
		return nil, err
	}

	return rsp.Payload, nil
}

func (api *grafanaAPI) DeleteFolder(ctx context.Context, uid string) (string, error) {
	params := &folders.DeleteFolderParams{
		FolderUID:  uid,
		Context:    ctx,
		HTTPClient: api.hc,
	}

	rsp, err := api.gc.Folders.DeleteFolder(params)
	if err != nil {
		return "", err
	}

	return *rsp.Payload.Title, nil
}

func (api *grafanaAPI) GetDashboardByUID(ctx context.Context, uid string) ([]byte, error) {
	params := &dashboards.GetDashboardByUIDParams{
		UID:        uid,
		Context:    ctx,
		HTTPClient: api.hc,
	}

	rsp, err := api.gc.Dashboards.GetDashboardByUIDWithParams(params)
	if err != nil {
		return nil, err
	}

	return rsp.Payload.MarshalBinary()
}

func (api *grafanaAPI) UpsertDashboard(ctx context.Context, folderUID string, dash any) ([]byte, error) {
	params := &dashboards.PostDashboardParams{
		Body: &models.SaveDashboardCommand{
			UpdatedAt: strfmt.DateTime{},
			Dashboard: dash,
			FolderUID: folderUID,
		},
		Context:    ctx,
		HTTPClient: api.hc,
	}

	rsp, err := api.gc.Dashboards.PostDashboardWithParams(params)
	if err != nil {
		return nil, err
	}

	data, err := rsp.Payload.MarshalBinary()
	return data, err
}

func (api *grafanaAPI) DeleteDashboard(ctx context.Context, uid string) (string, error) {
	params := &dashboards.DeleteDashboardByUIDParams{
		UID:        uid,
		Context:    ctx,
		HTTPClient: api.hc,
	}
	rsp, err := api.gc.Dashboards.DeleteDashboardByUIDWithParams(params)
	if err != nil {
		return "", err
	}
	return *rsp.Payload.Title, nil
}
