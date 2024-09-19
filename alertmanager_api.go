package pag

import (
	"bytes"
	"context"
	"net/http"
	urlpkg "net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"sigs.k8s.io/yaml"
)

type AlertManagerYAML struct {
	Global *AlertManagerGlobalYAML `yaml:"global,omitempty"`

	Templates []string `yaml:"templates,omitempty"`

	Route AlertManagerRoute `yaml:"route,omitempty"`

	InhibitRule []AlertManagerInhibitRuleYAML `yaml:"inhibit_rule,omitempty"`
}

type AlertManagerGlobalYAML struct {
	SmtpFrom             string `yaml:"smtp_from,omitempty"`
	SmtpSmartHost        string `yaml:"smtp_smarthost,omitempty"`
	SmtpHello            string `yaml:"smtp_hello,omitempty"`
	SmtpAuthUsername     string `yaml:"smtp_auth_username,omitempty"`
	SmtpAuthPassword     string `yaml:"smtp_auth_password,omitempty"`
	SmtpAuthPasswordFile string `yaml:"smtp_auth_password_file,omitempty"`
	SmtpAuthIdentify     string `yaml:"smtp_auth_identity,omitempty"`
	SmtpAuthSecret       string `yaml:"smtp_auth_secret,omitempty"`
	SmtpRequireTLS       bool   `yaml:"smtp_require_tls,omitempty"`

	//SlackAPIURL     string `yaml:"slack_api_url,omitempty"`
	//SlackAPIURLFile string `yaml:"slack_api_url_file,omitempty"`
	//
	//VictoropsAPIKey     string `yaml:"victorops_api_key,omitempty"`
	//VictoropsAPIKeyFile string `yaml:"victorops_api_key_file,omitempty"`
	//VictoropsAPIURL     string `yaml:"victorops_api_url,omitempty"`
	//
	//PagerDutyURL string `yaml:"pagerduty_url,omitempty"`
	//
	//OpsGenieAPIKey     string `yaml:"opsgenie_api_key,omitempty"`
	//OpsGenieAPIKeyFile string `yaml:"opsgenie_api_key_file,omitempty"`
	//OpsGenieAPIURL     string `yaml:"opsgenie_api_url,omitempty"`

	WechatAPIURL    string `yaml:"wechat_api_url,omitempty"`
	WechatAPISecret string `yaml:"wechat_api_secret,omitempty"`
	WechatAPICorpID string `yaml:"wechat_api_corp_id,omitempty"`

	//TelegramAPIURL string `yaml:"telegram_api_url,omitempty"`
	//WebexAPIURL    string `yaml:"webex_api_url,omitempty"`

	HTTPConfig *config.HTTPClientConfig `yaml:"http_config,omitempty"`

	ResolveTimeout model.Duration `yaml:"resolve_timeout,omitempty"`
}

type AlertManagerRoute struct {
	Receiver string   `yaml:"receiver,omitempty"`
	GroupBy  []string `yaml:"group_by,omitempty"`
	Continue bool     `yaml:"continue,omitempty"`

	Match    map[string]string `yaml:"match,omitempty"`
	MatchRe  map[string]string `yaml:"match_re,omitempty"`
	Matchers []string          `yaml:"matchers,omitempty"`

	GroupWaits    model.Duration `yaml:"group_waits,omitempty"`
	GroupInterval model.Duration `yaml:"group_interval,omitempty"`

	RepeatInterval model.Duration `yaml:"repeat_interval,omitempty"`

	MuteTimeIntervals []string `yaml:"mute_time_intervals,omitempty"`

	ActiveTimeIntervals []string `yaml:"active_time_intervals,omitempty"`

	Routes []*AlertManagerRoute `yaml:"routes,omitempty"`
}

type AlertManagerReceiverYAML struct {
	Name string `yaml:"name"`

	EmailConfigs []ReceiverEmailConfig `yaml:"email_config,omitempty"`

	WebhookConfigs []ReceiverWebhookYAML `yaml:"webhook_configs,omitempty"`

	WeChatConfigs []ReceiverWechatYAML `yaml:"wechat_configs,omitempty"`
}

type AlertManagerInhibitRuleYAML struct {
	TargetMatch    map[string]string `yaml:"target_match,omitempty"`
	TargetMatchRe  map[string]string `yaml:"target_match_re,omitempty"`
	TargetMatchers []string          `yaml:"target_matchers,omitempty"`

	SourceMatch    map[string]string `yaml:"source_match,omitempty"`
	SourceMatchRe  map[string]string `yaml:"source_match_re,omitempty"`
	SourceMatchers []string          `yaml:"source_matchers,omitempty"`

	Equal []string `yaml:"equal,omitempty"`
}

type ReceiverEmailConfig struct {
	SendResolved bool   `yaml:"send_resolved,omitempty"`
	To           string `yaml:"to"`
	From         string `yaml:"from,omitempty"`
	SmartHost    string `yaml:"smart_host,omitempty"`
	Hello        string `yaml:"hello,omitempty"`

	AuthUsername     string `yaml:"auth_username,omitempty"`
	AuthPassword     string `yaml:"auth_password,omitempty"`
	AuthPasswordFile string `yaml:"auth_password_file,omitempty"`
	AuthSecret       string `yaml:"auth_secret,omitempty"`
	AuthIdentify     string `yaml:"auth_identity,omitempty"`

	RequiredTLS bool `yaml:"required_tls,omitempty"`

	TlsConfig *config.TLSConfig `yaml:"tls_config,omitempty"`

	Html string `yaml:"html,omitempty"`
	Text string `yaml:"text,omitempty"`

	Headers map[string]string `yaml:"headers,omitempty"`
}

type ReceiverWebhookYAML struct {
	SendResolved bool                     `yaml:"send_resolved,omitempty"`
	URL          string                   `yaml:"url,omitempty"`
	URLFile      string                   `yaml:"url_file,omitempty"`
	HTTPConfig   *config.HTTPClientConfig `yaml:"http_config,omitempty"`
	MaxAlerts    uint32                   `yaml:"max_alerts,omitempty"`
}

type ReceiverWechatYAML struct {
	SendResolved bool   `yaml:"send_resolved,omitempty"`
	APISecret    string `yaml:"api_secret,omitempty"`
	APIURL       string `yaml:"api_url,omitempty"`
	CorpID       string `yaml:"corp_id,omitempty"`
	Message      string `yaml:"message,omitempty"`
	MessageType  string `yaml:"message_type,omitempty"`
	AgentID      string `yaml:"agent_id,omitempty"`
	ToUser       string `yaml:"to_user,omitempty"`
	ToParty      string `yaml:"to_party,omitempty"`
	ToTag        string `yaml:"to_tag,omitempty"`
}

type AlertManagerAPI interface {
	ConfigYAML() AlertManagerYAML

	Healthy(ctx context.Context) error
	Ready(ctx context.Context) error
	Reload(ctx context.Context) error
}

func NewAlertManagerAPI(hc *http.Client, cfg *AlertManagerConfig) (AlertManagerAPI, error) {
	api := &alertManagerAPI{
		cfg: cfg,
		hc:  hc,
	}

	err := api.load()
	if err != nil {
		return nil, err
	}

	return api, nil
}

type alertManagerAPI struct {
	cfg *AlertManagerConfig

	hc *http.Client

	yml *AlertManagerYAML
}

func (api *alertManagerAPI) load() error {
	dst := filepath.Join(api.cfg.ConfigYAML)
	data, err := os.ReadFile(dst)
	if err != nil {
		return err
	}

	var yml AlertManagerYAML
	if err = yaml.Unmarshal(data, &yml); err != nil {
		return err
	}
	api.yml = &yml

	return nil
}

func (api *alertManagerAPI) ConfigYAML() AlertManagerYAML {
	var out AlertManagerYAML
	out = *api.yml
	*out.Global = *api.yml.Global
	return out
}

func (api *alertManagerAPI) endpoint() *urlpkg.URL {
	endpoint := api.cfg.Endpoint
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "http://" + endpoint
	}

	u, _ := urlpkg.Parse(endpoint)
	return u
}

func (api *alertManagerAPI) URL(ep string, args map[string]string) *urlpkg.URL {
	p := path.Join(api.endpoint().Path, ep)

	for arg, val := range args {
		arg = ":" + arg
		p = strings.ReplaceAll(p, arg, val)
	}

	u := *api.endpoint()
	u.Path = p

	return &u
}

func (api *alertManagerAPI) Do(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	resp, err := api.hc.Do(req)
	defer func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()

	if err != nil {
		return nil, nil, err
	}

	var body []byte
	done := make(chan struct{})
	go func() {
		var buf bytes.Buffer
		_, err = buf.ReadFrom(resp.Body)
		body = buf.Bytes()
		close(done)
	}()

	select {
	case <-ctx.Done():
		<-done
		err = resp.Body.Close()
		if err == nil {
			err = ctx.Err()
		}
	case <-done:
	}

	return resp, body, err
}

func (api *alertManagerAPI) Healthy(ctx context.Context) error {
	url := api.URL("/-/healthy", map[string]string{})

	req := &http.Request{
		Method: http.MethodGet,
		URL:    url,
	}
	_, _, err := api.Do(ctx, req)
	if err != nil {
		return err
	}
	return err
}

func (api *alertManagerAPI) Ready(ctx context.Context) error {
	url := api.URL("/-/ready", map[string]string{})

	req := &http.Request{
		Method: http.MethodGet,
		URL:    url,
	}
	_, _, err := api.Do(ctx, req)
	if err != nil {
		return err
	}
	return err
}

func (api *alertManagerAPI) Reload(ctx context.Context) error {
	url := api.URL("/-/reload", map[string]string{})

	req := &http.Request{
		Method: http.MethodPost,
		URL:    url,
	}
	_, _, err := api.Do(ctx, req)
	if err != nil {
		return err
	}
	return err
}

// func (api *alertManagerAPI)
