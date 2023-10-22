package main

import (
	"flag"
	"fmt"
	"github.com/csyezheng/ip-scanner/common"
	"github.com/csyezheng/ip-scanner/sites"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"time"
)

func main() {
	configFilePath := flag.String("config", "./configs/config.toml", "Config file, toml format")
	siteFlag := flag.String("site", "",
		"This option should specify the site that exists under Sites configured in config.toml, such as GoogleTranslate, Cloudflare")
	flag.Parse()
	viper.SetConfigType("toml")
	viper.SetConfigFile(*configFilePath)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var config common.Config
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}
	if *siteFlag != "" {
		config.General.Site = *siteFlag
	}
	config.Ping.Timeout = config.Ping.Timeout * time.Millisecond
	config.HTTP.Timeout = config.HTTP.Timeout * time.Millisecond
	if config.General.Debug {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
		logger := slog.New(handler)
		slog.SetDefault(logger)
	}
	switch config.General.Site {
	case "GoogleTranslate":
		err := sites.FetchGTIPRanges(&config)
		if err != nil {
			slog.Error("error occur %s", err)
		}
	case "Cloudflare":
		err := sites.FetchCFIPRanges(&config)
		if err != nil {
			slog.Error("error occur %s", err)
		}
	}
}
