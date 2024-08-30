package cmd

import (
	"context"

	"github.com/ksysoev/oneway/pkg/core/revconproxy"
	revsvc "github.com/ksysoev/oneway/pkg/svc/revconproxy"
)

var revConConfig = &revconproxy.Config{
	NameSpace: "example",
	CtrlAPI:   "exchange:9090",
	ConnAPI:   "exchange:9091",
	Services: []revconproxy.ServiceCongfig{
		{
			Name:    "echoserver",
			Address: "echoserver:9090",
		},
		{
			Name:    "restapi",
			Address: "httpserver:8080",
		},
	},
}

func runRevConProxy(ctx context.Context) error {
	svc := revconproxy.New(revConConfig)

	revproxy := revsvc.New(svc, revConConfig.CtrlAPI)

	return revproxy.Run(ctx)
}
