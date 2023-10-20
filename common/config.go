package common

import "time"

type Config struct {
	General struct {
		UsedFor      string
		Debug        bool
		Workers      int
		ScannedLimit int
		FoundLimit   int
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
