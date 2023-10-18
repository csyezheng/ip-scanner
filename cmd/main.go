package main

import (
	"flag"
	"fmt"
	"github.com/csyezheng/ip-scanner/common"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"time"
)

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	configFilePath := flag.String("config", "./configs/config.toml", "Config file, toml format")
	domain := flag.String("domain", "cloudflare", "domain: cloudflare or google")
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
	if isFlagPassed("domain") {
		config.General.Domain = *domain
	}

	config.Ping.Timeout = config.Ping.Timeout * time.Millisecond
	config.HTTP.Timeout = config.HTTP.Timeout * time.Millisecond
	if config.General.Debug {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
		logger := slog.New(handler)
		slog.SetDefault(logger)
	}
	common.Start(&config)
}
