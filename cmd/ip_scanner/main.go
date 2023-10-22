package main

import (
	"flag"
	"github.com/csyezheng/ip-scanner/cmd"
)

func main() {
	configFilePath := flag.String("config", "./configs/config.toml", "Config file, toml format")
	siteFlag := flag.String("site", "",
		"This option should specify the site that exists under Sites configured in config.toml, such as GoogleTranslate, Cloudflare")
	flag.Parse()
	cmd.Start(*configFilePath, *siteFlag)
}
