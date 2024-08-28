package cmd

import (
	"github.com/ksysoev/oneway/pkg/svc/connapi"
	"github.com/ksysoev/oneway/pkg/svc/ctrlapi"
	"github.com/ksysoev/oneway/pkg/svc/proxy"
)

type AppConfig struct {
	CtrlApi  *ctrlapi.Config
	ConnApi  *connapi.Config
	ProxyApi *proxy.Config
}

var appConfig = &AppConfig{
	CtrlApi: &ctrlapi.Config{
		Listen: ":9090",
	},
	ConnApi: &connapi.Config{
		Listen: ":9091",
	},
	ProxyApi: &proxy.Config{
		Listen: ":1080",
	},
}
