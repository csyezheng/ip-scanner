package common

import "time"

type Config struct {
	General struct {
		Site         string
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
	Sites struct {
		Cloudflare struct {
			IPRangesAPI        string
			IPRangesFile       string
			CustomIPRangesFile string
			IPOutputFile       string
			WithIPv6           bool
			HttpsURL           string
		}
		GoogleTranslate struct {
			IPRangesAPI        string
			IPRangesFile       string
			CustomIPRangesFile string
			IPOutputFile       string
			WithIPv6           bool
			HttpsURL           string
		}
	}
}
