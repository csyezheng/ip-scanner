package common

import (
	"bufio"
	"github.com/csyezheng/ip-scanner/usedfor"
	"log/slog"
	"net/netip"
	"os"
	"sync"
)

type IPArray struct {
	IPs      []string
	ipsMutex sync.Mutex
}

func (arr *IPArray) AddIP(ip string) {
	arr.ipsMutex.Lock()
	if arr.IPs == nil {
		arr.IPs = make([]string, 0)
	}
	arr.IPs = append(arr.IPs, ip)
	arr.ipsMutex.Unlock()
}

func CIDRToIPs(cidrAddress string, iparr *IPArray, wg *sync.WaitGroup) {
	defer wg.Done()
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
		iparr.AddIP(addr.String())
		addr = addr.Next()
	}
}

func GetIPs(config *Config) []string {
	var iparr IPArray
	var cidrs []string
	usedFor := config.General.UsedFor
	if usedFor == "Cloudflare" {
		var cloudflare usedfor.CloudFlare
		customIPRangesFile := config.UsedFor.Cloudflare.CustomIPRangesFile
		ipRangesFile := config.UsedFor.Cloudflare.IPRangesFile
		withIPv6 := config.UsedFor.Cloudflare.WithIPv6
		err := cloudflare.LoadCIDRs(customIPRangesFile, ipRangesFile, withIPv6)
		if err != nil {
			slog.Error("Loading CIDRs failed:", err)
		}
		cidrs = cloudflare.CIDRs

	} else if usedFor == "GoogleTranslate" {
		var googleTranslate usedfor.GoogleTranslate
		customIPRangesFile := config.UsedFor.GoogleTranslate.CustomIPRangesFile
		ipRangesFile := config.UsedFor.GoogleTranslate.IPRangesFile
		withIPv6 := config.UsedFor.Cloudflare.WithIPv6
		err := googleTranslate.LoadCIDRs(customIPRangesFile, ipRangesFile, withIPv6)
		if err != nil {
			slog.Error("Loading CIDRs failed:", err)
		}
		cidrs = googleTranslate.CIDRs
	}
	var wg sync.WaitGroup
	for _, cidrAddress := range cidrs {
		wg.Add(1)
		go CIDRToIPs(cidrAddress, &iparr, &wg)
	}
	wg.Wait()
	slog.Info("Load IPs:", "Count", len(iparr.IPs))
	return iparr.IPs
}

func writeToFile(scanRecords ScanRecordArray, config *Config) {
	usedFor := config.General.UsedFor
	var outputFile string
	if usedFor == "Cloudflare" {
		outputFile = config.UsedFor.Cloudflare.IPOutputFile
	} else if usedFor == "GoogleTranslate" {
		outputFile = config.UsedFor.GoogleTranslate.IPOutputFile
	}
	f, err := os.Create(outputFile)
	if err != nil {
		slog.Error("Failed to create file", err)
	}
	w := bufio.NewWriter(f)
	for _, record := range scanRecords {
		w.WriteString(record.IP + "\n")
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
				slog.String("Protocal", record.Protocol), slog.Float64("PingRTT", record.PingRTT),
				slog.Float64("HttpRTT", record.HttpRTT))
		}
	}
}
