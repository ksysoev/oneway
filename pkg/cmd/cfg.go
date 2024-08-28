package cmd

import (
	"os"

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
		Listen: os.Getenv("MANAGE_API"),
	},
	ConnApi: &connapi.Config{
		Listen: os.Getenv("CONNECTION_API"),
	},
	ProxyApi: &proxy.Config{
		Listen: os.Getenv("PROXY_SERVER"),
	},
}
