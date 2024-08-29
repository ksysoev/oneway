package cmd

import (
	"context"

	"github.com/ksysoev/oneway/pkg/svc/revconproxy"
)

func runRevConProxy(ctx context.Context) error {
	svc := revconproxy.New()

	return svc.Run(ctx)
}
