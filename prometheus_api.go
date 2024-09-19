package pag

import (
	"context"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"sigs.k8s.io/yaml"
)

type PrometheusYAML struct {
	Global PrometheusGlobalYAML `yaml:"global"`

	RuleFiles []string `yaml:"rule_files"`

	ScrapeConfigs []PrometheusScrapeConfigYAML `yaml:"scrape_configs"`
}

type PrometheusGlobalYAML struct {
	// How frequently to scrape targets by default.
	ScrapeInterval model.Duration `yaml:"scrape_interval,omitempty"`
	// The default timeout when scraping targets.
	ScrapeTimeout model.Duration `yaml:"scrape_timeout,omitempty"`
	// How frequently to evaluate rules by default.
	EvaluationInterval model.Duration `yaml:"evaluation_interval,omitempty"`
	// Offset the rule evaluation timestamp of this particular group by the
	// specified duration into the past to ensure the underlying metrics have been received.
	RuleQueryOffset model.Duration `yaml:"rule_query_offset,omitempty"`
	// File to which PromQL queries are logged.
	QueryLogFile string `yaml:"query_log_file,omitempty"`
	// File to which scrape failures are logged.
	ScrapeFailureLogFile string `yaml:"scrape_failure_log_file,omitempty"`
	// The labels to add to any timeseries that this Prometheus instance scrapes.
	ExternalLabels map[string]string `yaml:"external_labels,omitempty"`
}

type PrometheusScrapeConfigYAML struct {
	// The job name to which the job label is set by default.
	JobName string `yaml:"job_name"`
	// Indicator whether the scraped metrics should remain unmodified.
	HonorLabels bool `yaml:"honor_labels,omitempty"`
	// Indicator whether the scraped timestamps should be respected.
	HonorTimestamps bool `yaml:"honor_timestamps"`
	// Indicator whether to track the staleness of the scraped timestamps.
	TrackTimestampsStaleness bool `yaml:"track_timestamps_staleness"`
	// A set of query parameters with which the target is scraped.
	Params urlpkg.Values `yaml:"params,omitempty"`
	// How frequently to scrape the targets of this scrape config.
	ScrapeInterval model.Duration `yaml:"scrape_interval,omitempty"`
	// The timeout for scraping targets of this config.
	ScrapeTimeout model.Duration `yaml:"scrape_timeout,omitempty"`
	// Whether to scrape a classic histogram that is also exposed as a native histogram.
	ScrapeClassicHistograms bool `yaml:"scrape_classic_histograms,omitempty"`
	// File to which scrape failures are logged.
	ScrapeFailureLogFile string `yaml:"scrape_failure_log_file,omitempty"`
	// The HTTP resource path on which to fetch metrics from targets.
	MetricsPath string `yaml:"metrics_path,omitempty"`
	// The URL scheme with which to fetch metrics from targets.
	Scheme string `yaml:"scheme,omitempty"`
	// Indicator whether to request compressed response from the target.
	EnableCompression bool `yaml:"enable_compression"`

	// We cannot do proper Go type embedding below as the parser will then parse
	// values arbitrarily into the overflow maps of further-down types.

	HTTPClientConfig config.HTTPClientConfig `yaml:",inline"`

	StaticConfig []ServiceDiscoveryEndpoint `yaml:"static_configs,omitempty"`

	FileSDConfigs []FileSDConfig `yaml:"file_sd_configs,omitempty"`
}

type FileSDConfig struct {
	Files []string `yaml:"files"`

	RefreshInterval model.Duration `yaml:"refresh_interval,omitempty"`
}

type PrometheusAPI interface {
	ConfigYAML() PrometheusYAML
	Healthy(ctx context.Context) error
	Ready(ctx context.Context) error
	Reload(ctx context.Context) error

	Values(ctx context.Context) (model.LabelValues, error)

	Query(ctx context.Context, query string, ts time.Time, opts ...prometheusv1.Option) (model.Value, prometheusv1.Warnings, error)
	QueryRange(ctx context.Context, query string, rg prometheusv1.Range, opts ...prometheusv1.Option) (model.Value, prometheusv1.Warnings, error)
	QueryExemplars(ctx context.Context, query string, start, end time.Time) ([]prometheusv1.ExemplarQueryResult, error)

	AddTarget(ctx context.Context, sd *ServiceDiscovery) error
	Targets(ctx context.Context) (prometheusv1.TargetsResult, error)

	AddRuleGroups(ctx context.Context, rg *RuleGroup) error
	GetRules(ctx context.Context) (prometheusv1.RulesResult, error)

	Alerts(ctx context.Context) (prometheusv1.AlertsResult, error)
}

// @TIP: hello
func NewPrometheusAPI(hc *http.Client, cfg *PrometheusConfig) (PrometheusAPI, error) {
	address := cfg.Endpoint
	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}

	apiCfg := api.Config{
		Address: address,
		Client:  hc,
	}
	pc, err := api.NewClient(apiCfg)
	if err != nil {
		return nil, err
	}

	pa := &prometheusAPI{
		cfg: cfg,
		c:   pc,
	}

	if err = pa.load(); err != nil {
		return nil, err
	}

	return pa, nil
}

type prometheusAPI struct {
	cfg *PrometheusConfig

	py *PrometheusYAML

	c api.Client
}

func (pa *prometheusAPI) load() error {
	data, err := os.ReadFile(filepath.Join(pa.cfg.ConfigYAML))
	if err != nil {
		return err
	}
	var py PrometheusYAML
	if err = yaml.Unmarshal(data, &py); err != nil {
		return err
	}
	pa.py = &py

	return nil
}

func (pa *prometheusAPI) newAPI() prometheusv1.API {
	return prometheusv1.NewAPI(pa.c)
}

func (pa *prometheusAPI) ConfigYAML() PrometheusYAML {
	return *pa.py
}

func (pa *prometheusAPI) Healthy(ctx context.Context) error {
	url := pa.c.URL("/-/healthy", map[string]string{})

	req := &http.Request{
		Method: http.MethodGet,
		URL:    url,
	}
	_, _, err := pa.c.Do(ctx, req)
	if err != nil {
		return err
	}
	return err
}

func (pa *prometheusAPI) Ready(ctx context.Context) error {
	url := pa.c.URL("/-/ready", map[string]string{})

	req := &http.Request{
		Method: http.MethodGet,
		URL:    url,
	}
	_, _, err := pa.c.Do(ctx, req)
	if err != nil {
		return err
	}
	return err
}

func (pa *prometheusAPI) Reload(ctx context.Context) error {
	url := pa.c.URL("/-/reload", map[string]string{})

	req := &http.Request{
		Method: http.MethodPost,
		URL:    url,
	}
	_, _, err := pa.c.Do(ctx, req)
	if err != nil {
		return err
	}
	return err
}

func (pa *prometheusAPI) Values(ctx context.Context) (model.LabelValues, error) {
	values, _, err := pa.newAPI().LabelValues(ctx, "__name__", nil, time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (pa *prometheusAPI) Query(ctx context.Context, query string, ts time.Time, opts ...prometheusv1.Option) (model.Value, prometheusv1.Warnings, error) {
	return pa.newAPI().Query(ctx, query, ts, opts...)
}

func (pa *prometheusAPI) QueryRange(ctx context.Context, query string, rg prometheusv1.Range, opts ...prometheusv1.Option) (model.Value, prometheusv1.Warnings, error) {
	return pa.newAPI().QueryRange(ctx, query, rg, opts...)
}

func (pa *prometheusAPI) QueryExemplars(ctx context.Context, query string, start, end time.Time) ([]prometheusv1.ExemplarQueryResult, error) {
	return pa.newAPI().QueryExemplars(ctx, query, start, end)
}

func (pa *prometheusAPI) AddTarget(ctx context.Context, sd *ServiceDiscovery) error {
	var dir string
	for _, sc := range pa.py.ScrapeConfigs {
		if len(sc.FileSDConfigs) != 0 && len(sc.FileSDConfigs) != 0 {
			dir = filepath.Dir(sc.FileSDConfigs[0].Files[0])
			break
		}
	}

	if dir == "" {
		return fmt.Errorf("no file_sd_configs configured")
	}

	dst := filepath.Join(dir, sd.Name+".yaml")
	data, err := yaml.Marshal(sd.Endpoints)
	if err != nil {
		return err
	}
	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (pa *prometheusAPI) Targets(ctx context.Context) (prometheusv1.TargetsResult, error) {
	return pa.newAPI().Targets(ctx)
}

func (pa *prometheusAPI) AddRuleGroups(ctx context.Context, rg *RuleGroup) error {
	if len(pa.py.RuleFiles) == 0 {
		return fmt.Errorf("rules directory not exists")
	}

	rd := filepath.Dir(pa.py.RuleFiles[0])
	dst := filepath.Join(rd, rg.Name+".yml")
	groups := &RuleGroups{
		Groups: []RuleGroup{*rg},
	}
	data, err := yaml.Marshal(groups)
	if err != nil {
		return err
	}
	if err = os.WriteFile(dst, data, 0644); err != nil {
		return err
	}

	return pa.Reload(ctx)
}

func (pa *prometheusAPI) GetRules(ctx context.Context) (prometheusv1.RulesResult, error) {
	return pa.newAPI().Rules(ctx)
}

func (pa *prometheusAPI) Alerts(ctx context.Context) (prometheusv1.AlertsResult, error) {
	return pa.newAPI().Alerts(ctx)
}
