package sites

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type GoogleTranslate struct {
	IPs   []string
	CIDRs []string
}

type gtRsponse struct {
	SyncToken    string `json:"syncToken"`
	CreationTime string `json:"creationTime"`
	Prefixes     []struct {
		Ipv4Prefix string `json:"ipv4Prefix,omitempty"`
		Ipv6Prefix string `json:"ipv6Prefix,omitempty"`
	} `json:"prefixes"`
}

func FetchGTIPRanges(url string, dest string) error {
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
	var res gtRsponse
	if err = json.Unmarshal(body, &res); err != nil {
		slog.Error("Failed to decode release JSON. Error:", err)
		return err
	}
	var ipCIDRs []string
	items := res.Prefixes
	for _, item := range items {
		ipv4CIDR := item.Ipv4Prefix
		ipv6CIDR := item.Ipv6Prefix
		if ipv4CIDR != "" {
			ipCIDRs = append(ipCIDRs, ipv4CIDR)
		}
		if ipv6CIDR != "" {
			ipCIDRs = append(ipCIDRs, ipv6CIDR)
		}
	}
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	_, err = f.WriteString(strings.Join(ipCIDRs, "\n"))
	if err != nil {
		return err
	}
	slog.Info("fetch success, the latest IP segment has been saved in", "file", dest)
	return nil
}
