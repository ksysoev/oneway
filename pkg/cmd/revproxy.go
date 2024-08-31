package cmd

import (
	"context"

	"github.com/ksysoev/oneway/pkg/core/revconproxy"
	revsvc "github.com/ksysoev/oneway/pkg/svc/revconproxy"
)

func runRevProxy(ctx context.Context, cfg *revconproxy.Config) error {
	svc := revconproxy.New(cfg)

	revproxy := revsvc.New(svc, cfg.CtrlAPI)

	return revproxy.Run(ctx)
}
