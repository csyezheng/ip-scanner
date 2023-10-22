package cmd

import (
	"fmt"
	"github.com/csyezheng/ip-scanner/common"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"time"
)

func Start(configFilePath string, siteFlag string) {
	viper.SetConfigType("toml")
	viper.SetConfigFile(configFilePath)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var config common.Config
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}
	if siteFlag != "" {
		config.General.Site = siteFlag
	}
	config.Ping.Timeout = config.Ping.Timeout * time.Millisecond
	config.HTTP.Timeout = config.HTTP.Timeout * time.Millisecond
	if config.General.Debug {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
		logger := slog.New(handler)
		slog.SetDefault(logger)
	}
	common.Run(&config)
}
