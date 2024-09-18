package pag

type ServiceDiscoveryEndpoint struct {
	Targets []string          `json:"targets" yaml:"targets"`
	Labels  map[string]string `json:"labels" yaml:"labels"`
}

type ServiceDiscovery struct {
	Name string `json:"name"`

	Endpoints []ServiceDiscoveryEndpoint `json:"endpoints"`
}

type Rule struct {
	Alert       string            `json:"alert" yaml:"alert"`
	Expr        string            `json:"expr" yaml:"expr"`
	For         string            `json:"for" yaml:"for"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

type RuleGroup struct {
	Name  string `json:"name" yaml:"name"`
	Rules []Rule `json:"rules" yaml:"rules"`
}

type RuleGroups struct {
	Groups []RuleGroup `json:"groups" yaml:"groups"`
}
