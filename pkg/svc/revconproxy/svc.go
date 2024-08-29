package revconproxy

import "context"

type ServiceCongfig struct {
	Name    string
	Address string
}

type Config struct {
	NameSpace string
	CtrlAPI   string
	ConnAPI   string
	Services  []ServiceCongfig
}

type Proxy struct {
	config *Config
}

func New(cfg *Config) *Proxy {
	return &Proxy{
		config: cfg,
	}
}

func (s *Proxy) Run(ctx context.Context) error {
	return nil
}
