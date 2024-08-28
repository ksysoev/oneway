package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/ksysoev/oneway/pkg/core/exchange"
	"github.com/ksysoev/oneway/pkg/repo"
	"github.com/ksysoev/oneway/pkg/svc/connapi"
	"github.com/ksysoev/oneway/pkg/svc/ctrlapi"
	"github.com/ksysoev/oneway/pkg/svc/proxy"
)

func runExchange(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	cancel()

	connQueue := repo.NewConnectionQueue()
	revProxyRegistry := repo.NewRevProxyRegistry()

	exchange := exchange.New(revProxyRegistry, connQueue)

	ctrlAPI := ctrlapi.New(appConfig.CtrlApi, exchange)
	connApi := connapi.New(appConfig.ConnApi, exchange)
	sock5 := proxy.New(appConfig.ProxyApi, exchange)

	errs := make(chan error, 3)

	go func() { errs <- ctrlAPI.Run(ctx) }()
	go func() { errs <- connApi.Run(ctx) }()
	go func() { errs <- sock5.Run(ctx) }()

	return collectErrs(errs, 3)
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
