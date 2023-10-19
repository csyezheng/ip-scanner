package common

import "time"

type Config struct {
	General struct {
		Debug   bool
		Workers int
		UsedFor string
	}
	Ping struct {
		Protocol string
		Port     uint16
		Count    int
		Timeout  time.Duration
		all      bool
	}
	HTTP struct {
		Port    uint16
		Count   int
		Timeout time.Duration
		all     bool
	}
	UsedFor struct {
		Cloudflare struct {
			IPRangesFile       string
			CustomIPRangesFile string
			IPOutputFile       string
			WithIPv6           bool
			HttpsURL           string
		}
		GoogleTranslate struct {
			IPRangesFile       string
			CustomIPRangesFile string
			IPOutputFile       string
			WithIPv6           bool
			HttpsURL           string
		}
	}
}
