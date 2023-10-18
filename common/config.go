package common

import "time"

type Config struct {
	General struct {
		Debug   bool
		Workers int
		Domain  string
	}
	Ping struct {
		Protocol string
		Port     uint16
		Count    int
		Timeout  time.Duration
	}
	HTTP struct {
		Port    uint16
		Timeout time.Duration
	}
	Domains struct {
		Cloudflare struct {
			IPRangesFile string
			IPOutputFile string
			WithIPv6     bool
			HttpsURL     string
		}
		GoogleTranslate struct {
			IPRangesFile string
			IPOutputFile string
			WithIPv6     bool
			HttpsURL     string
		}
	}
}
