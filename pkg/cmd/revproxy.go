package cmd

import (
	"context"

	"github.com/ksysoev/oneway/pkg/core/revconproxy"
	"github.com/ksysoev/oneway/pkg/prov/bridge"
	revsvc "github.com/ksysoev/oneway/pkg/svc/revconproxy"
)

type RevProxyConfig struct {
	ConnApi *bridge.Config `mapstructure:"conn_api"`
	Service revconproxy.Config
}

func runRevProxy(ctx context.Context, cfg *RevProxyConfig) error {
	bridgeProvider := bridge.New(cfg.ConnApi)

	svc := revconproxy.New(&cfg.Service, bridgeProvider)

	revproxy := revsvc.New(svc, cfg.Service.CtrlAPI)

	return revproxy.Run(ctx)
}
