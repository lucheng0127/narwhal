package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	flags "github.com/jessevdk/go-flags"
	"github.com/lucheng0127/narwhal/internal/pkg/config"
	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
	"github.com/lucheng0127/narwhal/internal/pkg/version"
	"github.com/lucheng0127/narwhal/pkg/proxy"
	"github.com/sirupsen/logrus"
)

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Parse command line arguments
	var opts struct {
		ConfigFile string `short:"f" long:"config-file" description:"config file" default:"/etc/narwhal/config.yaml"`
		ConfigType string `short:"t" long:"config-type" description:"config file type(toml, yaml, json)" default:"yaml"`
		LogLevel   string `short:"l" long:"log-level" description:"log level"`
		Version    bool   `long:"version" description:"show version info"`
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Println("Narwhal version ", version.Version())
		os.Exit(0)
	}

	// Set log
	switch opts.LogLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	ctx := utils.NewTraceContext()
	// Parse config file
	conf, err := config.ReadConfigFile(opts.ConfigFile, opts.ConfigType)
	if err != nil {
		logger.Error(ctx, err.Error())
		os.Exit(1)
	}

	// Launch server
	var s proxy.Server = proxy.NewProxyServer(proxy.ListenPort(conf.Port))
	go s.Launch()
	logger.Info(ctx, "Narwhal server started")

	// Exist with signal
	<-sigCh
	stopServer(ctx, s)
}

func stopServer(ctx context.Context, s proxy.Server) {
	logger.Info(ctx, "Stopping narwhal server")
	s.Stop()
}
