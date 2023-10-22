package main

import (
	"flag"
	"github.com/csyezheng/ip-scanner/cmd"
)

func main() {
	configFilePath := flag.String("config", "./configs/config.toml", "Config file, toml format")
	flag.Parse()
	cmd.Start(*configFilePath, "Cloudflare")
}
