package common

import "time"

type Config struct {
	Protocol        string
	Port            float64
	Count           int
	Timeout         time.Duration
	Workers         int
	CloudflareAPI   string
	CloudflareToken string
}
