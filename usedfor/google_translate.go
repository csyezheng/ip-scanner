package usedfor

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type GoogleTranslate struct {
	IPs   []string
	CIDRs []string
}

type response struct {
	SyncToken    string `json:"syncToken"`
	CreationTime string `json:"creationTime"`
	Prefixes     []struct {
		Ipv4Prefix string `json:"ipv4Prefix,omitempty"`
		Ipv6Prefix string `json:"ipv6Prefix,omitempty"`
	} `json:"prefixes"`
}

func (gg *GoogleTranslate) LoadCIDRs(customIPRangesFile string, ipRangesFile string, withIPv6 bool) error {
	_, err := os.Stat(customIPRangesFile)
	if err == nil {
		bytesRead, _ := os.ReadFile(customIPRangesFile)
		fileContent := string(bytesRead)
		lines := strings.Split(fileContent, "\n")
		gg.CIDRs = append(gg.CIDRs, lines...)
		return nil
	} else if os.IsNotExist(err) {
		f, err := os.Open(ipRangesFile)
		if err != nil {
			slog.Error("Could not open ip address ranges file:", ipRangesFile)
			os.Exit(1)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {

			}
		}(f)

		var res response
		decoder := json.NewDecoder(f)
		if err := decoder.Decode(&res); err != nil {
			slog.Error("Failed to decode release JSON. Error:", err)
			return err
		}
		for _, v := range res.Prefixes {
			gg.CIDRs = append(gg.CIDRs, v.Ipv4Prefix)
			if withIPv6 {
				gg.CIDRs = append(gg.CIDRs, v.Ipv6Prefix)
			}
		}
		return nil
	} else {
		return fmt.Errorf("file %s stat error: %v", customIPRangesFile, err)
	}
}
