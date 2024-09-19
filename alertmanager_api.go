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
	Global *AlertManagerGlobalYAML `json:"global,omitempty"`

	Templates []string `json:"templates,omitempty"`

	Route AlertManagerRoute `json:"route,omitempty"`

	Receivers []AlertManagerReceiverYAML `json:"receivers,omitempty"`

	InhibitRule []AlertManagerInhibitRuleYAML `json:"inhibit_rule,omitempty"`
}

type AlertManagerGlobalYAML struct {
	SmtpFrom             string `json:"smtp_from,omitempty"`
	SmtpSmartHost        string `json:"smtp_smarthost,omitempty"`
	SmtpHello            string `json:"smtp_hello,omitempty"`
	SmtpAuthUsername     string `json:"smtp_auth_username,omitempty"`
	SmtpAuthPassword     string `json:"smtp_auth_password,omitempty"`
	SmtpAuthPasswordFile string `json:"smtp_auth_password_file,omitempty"`
	SmtpAuthIdentify     string `json:"smtp_auth_identity,omitempty"`
	SmtpAuthSecret       string `json:"smtp_auth_secret,omitempty"`
	SmtpRequireTLS       bool   `json:"smtp_require_tls,omitempty"`

	//SlackAPIURL     string `json:"slack_api_url,omitempty"`
	//SlackAPIURLFile string `json:"slack_api_url_file,omitempty"`
	//
	//VictoropsAPIKey     string `json:"victorops_api_key,omitempty"`
	//VictoropsAPIKeyFile string `json:"victorops_api_key_file,omitempty"`
	//VictoropsAPIURL     string `json:"victorops_api_url,omitempty"`
	//
	//PagerDutyURL string `json:"pagerduty_url,omitempty"`
	//
	//OpsGenieAPIKey     string `json:"opsgenie_api_key,omitempty"`
	//OpsGenieAPIKeyFile string `json:"opsgenie_api_key_file,omitempty"`
	//OpsGenieAPIURL     string `json:"opsgenie_api_url,omitempty"`

	WechatAPIURL    string `json:"wechat_api_url,omitempty"`
	WechatAPISecret string `json:"wechat_api_secret,omitempty"`
	WechatAPICorpID string `json:"wechat_api_corp_id,omitempty"`

	//TelegramAPIURL string `json:"telegram_api_url,omitempty"`
	//WebexAPIURL    string `json:"webex_api_url,omitempty"`

	HTTPConfig *config.HTTPClientConfig `json:"http_config,omitempty"`

	ResolveTimeout model.Duration `json:"resolve_timeout,omitempty"`
}

type AlertManagerRoute struct {
	Receiver string   `json:"receiver,omitempty"`
	GroupBy  []string `json:"group_by,omitempty"`
	Continue bool     `json:"continue,omitempty"`

	Match    map[string]string `json:"match,omitempty"`
	MatchRe  map[string]string `json:"match_re,omitempty"`
	Matchers []string          `json:"matchers,omitempty"`

	GroupWaits    model.Duration `json:"group_waits,omitempty"`
	GroupInterval model.Duration `json:"group_interval,omitempty"`

	RepeatInterval model.Duration `json:"repeat_interval,omitempty"`

	MuteTimeIntervals []string `json:"mute_time_intervals,omitempty"`

	ActiveTimeIntervals []string `json:"active_time_intervals,omitempty"`

	Routes []*AlertManagerRoute `json:"routes,omitempty"`
}

type AlertManagerReceiverYAML struct {
	Name string `json:"name"`

	EmailConfigs []ReceiverEmailConfig `json:"email_config,omitempty"`

	WebhookConfigs []ReceiverWebhookYAML `json:"webhook_configs,omitempty"`

	WeChatConfigs []ReceiverWechatYAML `json:"wechat_configs,omitempty"`
}

type AlertManagerInhibitRuleYAML struct {
	TargetMatch    map[string]string `json:"target_match,omitempty"`
	TargetMatchRe  map[string]string `json:"target_match_re,omitempty"`
	TargetMatchers []string          `json:"target_matchers,omitempty"`

	SourceMatch    map[string]string `json:"source_match,omitempty"`
	SourceMatchRe  map[string]string `json:"source_match_re,omitempty"`
	SourceMatchers []string          `json:"source_matchers,omitempty"`

	Equal []string `json:"equal,omitempty"`
}

type ReceiverEmailConfig struct {
	SendResolved bool   `json:"send_resolved,omitempty"`
	To           string `json:"to"`
	From         string `json:"from,omitempty"`
	SmartHost    string `json:"smart_host,omitempty"`
	Hello        string `json:"hello,omitempty"`

	AuthUsername     string `json:"auth_username,omitempty"`
	AuthPassword     string `json:"auth_password,omitempty"`
	AuthPasswordFile string `json:"auth_password_file,omitempty"`
	AuthSecret       string `json:"auth_secret,omitempty"`
	AuthIdentify     string `json:"auth_identity,omitempty"`

	RequiredTLS bool `json:"required_tls,omitempty"`

	TlsConfig *config.TLSConfig `json:"tls_config,omitempty"`

	Html string `json:"html,omitempty"`
	Text string `json:"text,omitempty"`

	Headers map[string]string `json:"headers,omitempty"`
}

type ReceiverWebhookYAML struct {
	SendResolved bool                     `json:"send_resolved,omitempty"`
	URL          string                   `json:"url,omitempty"`
	URLFile      string                   `json:"url_file,omitempty"`
	HTTPConfig   *config.HTTPClientConfig `json:"http_config,omitempty"`
	MaxAlerts    uint32                   `json:"max_alerts,omitempty"`
}

type ReceiverWechatYAML struct {
	SendResolved bool   `json:"send_resolved,omitempty"`
	APISecret    string `json:"api_secret,omitempty"`
	APIURL       string `json:"api_url,omitempty"`
	CorpID       string `json:"corp_id,omitempty"`
	Message      string `json:"message,omitempty"`
	MessageType  string `json:"message_type,omitempty"`
	AgentID      string `json:"agent_id,omitempty"`
	ToUser       string `json:"to_user,omitempty"`
	ToParty      string `json:"to_party,omitempty"`
	ToTag        string `json:"to_tag,omitempty"`
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
	if api.yml.Global != nil {
		out.Global = &AlertManagerGlobalYAML{}
		*out.Global = *api.yml.Global
	}
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
