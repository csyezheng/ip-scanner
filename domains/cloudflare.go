package domains

import (
	"encoding/json"
	"log/slog"
	"os"
)

type CloudFlare struct {
	IPs   []string
	CIDRs []string
}

type Response struct {
	Result struct {
		Ipv4Cidrs []string `json:"ipv4_cidrs"`
		Ipv6Cidrs []string `json:"ipv6_cidrs"`
		Etag      string   `json:"etag"`
	} `json:"result"`
	Success  bool  `json:"success"`
	Errors   []any `json:"errors"`
	Messages []any `json:"messages"`
}

func (cf *CloudFlare) LoadCIDRs(filepath string, withIPv6 bool) error {
	f, err := os.Open(filepath)
	if err != nil {
		slog.Error("Could not open ip address ranges file:", filepath)
		os.Exit(1)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	var response Response
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&response); err != nil {
		slog.Error("Failed to decode release JSON. Error:", err)
		return err
	}
	ipv4 := response.Result.Ipv4Cidrs
	ipv6 := response.Result.Ipv6Cidrs
	cf.CIDRs = append(cf.CIDRs, ipv4[:1]...)
	if withIPv6 {
		cf.CIDRs = append(cf.CIDRs, ipv6...)
	}
	return nil
}
