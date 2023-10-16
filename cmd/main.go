package main

import (
	"flag"
	"fmt"
	"github.com/csyezheng/ip-scanner/common"
	"github.com/spf13/viper"
	"log"
	"time"
)

func main() {
	filePath := flag.String("config", "./configs/config.toml", "Config file, toml format")
	flag.Parse()
	viper.SetConfigType("toml")
	viper.SetConfigFile(*filePath)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var config common.Config
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}
	config.Timeout = config.Timeout * time.Millisecond

	records := common.BatchTest(config)
	for i := 0; i < len(records); i++ {
		log.Println(records[i].Latency)
		log.Println(records[i].Success)
	}
}
