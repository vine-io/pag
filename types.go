package pag

type ServiceDiscoveryEndpoint struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

type ServiceDiscovery struct {
	Name string `json:"name"`

	Endpoints []ServiceDiscoveryEndpoint `json:"endpoints"`
}

type Rule struct {
	Alert       string            `json:"alert"`
	Expr        string            `json:"expr"`
	For         string            `json:"for"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type RuleGroup struct {
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}

type RuleGroups struct {
	Groups []RuleGroup `json:"groups"`
}
