package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/ksysoev/oneway/pkg/core/exchange"
	"github.com/ksysoev/oneway/pkg/repo"
	"github.com/ksysoev/oneway/pkg/svc/ctrlapi"
	"github.com/ksysoev/oneway/pkg/svc/proxy"
	"github.com/ksysoev/oneway/pkg/svc/revconnapi"
)

type ExchaneConfig struct {
	CtrlAPI  *ctrlapi.Config    `mapstructure:"ctrl_api"`
	ConnAPI  *revconnapi.Config `mapstructure:"conn_api"`
	ProxyAPI *proxy.Config      `mapstructure:"proxy_server"`
}

func runExchange(ctx context.Context, cfg *ExchaneConfig) error {
	ctx, cancel := context.WithCancel(ctx)

	connQueue := repo.NewConnectionQueue()
	revProxyRegistry := repo.NewRevProxyRegistry()

	exchangeSvc := exchange.New(revProxyRegistry, connQueue)

	ctrlAPI := ctrlapi.New(cfg.CtrlAPI, exchangeSvc)
	connAPI := revconnapi.New(cfg.ConnAPI, exchangeSvc)
	sock5 := proxy.New(cfg.ProxyAPI, exchangeSvc)

	const expectedErrs = 3
	errs := make(chan error, expectedErrs)

	go func() {
		defer cancel()
		errs <- ctrlAPI.Run(ctx)
	}()
	go func() {
		defer cancel()
		errs <- connAPI.Run(ctx)
	}()
	go func() {
		defer cancel()
		errs <- sock5.Run(ctx)
	}()

	return collectErrs(errs, expectedErrs)
}

func collectErrs(errs <-chan error, n int) error {
	collectedErrs := make([]error, 0, n)

	for i := 0; i < n; i++ {
		err := <-errs

		if err != nil {
			collectedErrs = append(collectedErrs, err)
		}
	}

	if len(collectedErrs) > 0 {
		return fmt.Errorf("failed to run exchange: %w", errors.Join(collectedErrs...))
	}

	return nil
}
