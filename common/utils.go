package common

import (
	"bufio"
	"github.com/csyezheng/ip-scanner/domains"
	"log/slog"
	"net/netip"
	"os"
)

func CIDRToIPs(cidrAddress string) []string {
	var ips []string
	p, err := netip.ParsePrefix(cidrAddress)
	if err != nil {
		slog.Error("invalid cidr:", slog.String("CIDR", cidrAddress), slog.Any("Error", err))
	}
	p = p.Masked()
	addr := p.Addr()
	for {
		if !p.Contains(addr) {
			break
		}
		slog.Debug(addr.String())
		ips = append(ips, addr.String())
		addr = addr.Next()
	}
	return ips
}

func GetIPs(config *Config) []string {
	var ips []string
	domain := config.General.Domain
	if domain == "cloudflare" {
		var cloudflare domains.CloudFlare
		inputFile := config.Domains.Cloudflare.IPRangesFile
		withIPv6 := config.Domains.Cloudflare.WithIPv6
		err := cloudflare.LoadCIDRs(inputFile, withIPv6)
		if err != nil {
			slog.Error("Loading CIDRs failed:", err)
		}
		for _, cidrAddress := range cloudflare.CIDRs {
			ipList := CIDRToIPs(cidrAddress)
			ips = append(ips, ipList...)
		}
	} else if domain == "google" {
		var google domains.Google
		inputFile := config.Domains.GoogleTranslate.IPRangesFile
		withIPv6 := config.Domains.Cloudflare.WithIPv6
		err := google.LoadCIDRs(inputFile, withIPv6)
		if err != nil {
			slog.Error("Loading CIDRs failed:", err)
		}
		for _, cidrAddress := range google.CIDRs {
			ipList := CIDRToIPs(cidrAddress)
			ips = append(ips, ipList...)
		}
	}
	return ips
}

func writeToFile(scanRecords ScanRecordArray, config *Config) {
	domain := config.General.Domain
	var outputFile string
	if domain == "cloudflare" {
		outputFile = config.Domains.Cloudflare.IPOutputFile
	} else if domain == "google" {
		outputFile = config.Domains.Cloudflare.IPOutputFile
	}
	f, err := os.Create(outputFile)
	if err != nil {
		slog.Error("Failed to create file", err)
	}
	w := bufio.NewWriter(f)
	for _, record := range scanRecords {
		w.WriteString(record.IP)
	}
	w.Flush()
}

func printResult(scanRecords ScanRecordArray) {
	if len(scanRecords) == 0 {
		slog.Info("No found available ip!")
		return
	}
	for i, record := range scanRecords {
		if i < 10 {
			slog.Info("Scan Result:", slog.String("IP", record.IP),
				slog.String("Protocal", record.Protocol), slog.Float64("Latency", record.Latency))
		}
	}
}
