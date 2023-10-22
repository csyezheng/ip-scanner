package sites

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type cfResponse struct {
	Result struct {
		Ipv4Cidrs []string `json:"ipv4_cidrs"`
		Ipv6Cidrs []string `json:"ipv6_cidrs"`
		Etag      string   `json:"etag"`
	} `json:"result"`
	Success  bool  `json:"success"`
	Errors   []any `json:"errors"`
	Messages []any `json:"messages"`
}

func FetchCFIPRanges(url string, dest string) error {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("error on response.\n[ERROR] -", err)
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error while reading the response bytes:", err)
		return err
	}
	var res cfResponse
	if err = json.Unmarshal(body, &res); err != nil {
		slog.Error("Failed to decode release JSON. Error:", err)
		return err
	}
	if res.Success {
		ipv4CIDRs := res.Result.Ipv4Cidrs
		ipv6CIDRs := res.Result.Ipv6Cidrs
		ipCIDRs := append(ipv4CIDRs, ipv6CIDRs...)
		f, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		_, err = f.WriteString(strings.Join(ipCIDRs, "\n"))
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Fatch failed, cloudflare return: %v", res)
	}
	slog.Info("fetch success, the latest IP segment has been saved in", "file", dest)
	return nil
}
