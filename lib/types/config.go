package types

type Route struct {
	Path   string   `yaml:"path"`
	Hops   []string `yaml:"hops"`
	Target string   `yaml:"target"`
}

type Host struct {
	Host   string   `yaml:"host"`
	Routes []*Route `yaml:"routes"`
}

type Config struct {
	Addr       string   `json:"addr" yaml:"addr"`
	Acme       bool     `json:"acme" yaml:"acme"`
	Debug      bool     `json:"debug" yaml:"debug"`
	Certs      string   `json:"certs" yaml:"certs"`
	Hops       []string `yaml:"hops"`
	Hosts      []*Host  `json:"hosts" yaml:"hosts"`
	Https      bool     `json:"https" yaml:"https"`
	SelfSigned bool     `json:"self_signed" yaml:"self_signed"`
	Cert       string   `json:"cert" yaml:"cert"`
	Key        string   `json:"key" yaml:"key"`
}
