package common

import (
	"bufio"
	"fmt"
	"github.com/csyezheng/ip-scanner/usedfor"
	"io"
	"log/slog"
	"net/netip"
	"os"
	"runtime"
	"strings"
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
		_, err := w.WriteString(record.IP + "\n")
		if err != nil {
			slog.Error("write to output file failed", "error", err)
		}
	}
	err = w.Flush()
	if err != nil {
		slog.Error("flush failed", "error", err)
	}
}

func printResult(scanRecords ScanRecordArray, config *Config) {
	if len(scanRecords) == 0 {
		slog.Info("No found available ip!")
		return
	}
	head := scanRecords
	if len(head) > 10 {
		head = head[:10]
	}
	fmt.Printf("%s\t%s\t%s\t%s\n", "IP", "Protocol", "PingRTT", "HttpRTT")
	for _, record := range head {
		fmt.Printf("%s\t%s\t%f\t%f\n", record.IP, record.Protocol, record.PingRTT, record.HttpRTT)
	}
	if config.General.UsedFor == "GoogleTranslate" {
		fastestRecord := *scanRecords[0]
		slog.Info("The fastest IP has been found:")
		fmt.Printf("%v\t%s\n", fastestRecord.IP, "translate.googleapis.com")
		fmt.Printf("%v\t%s\n", fastestRecord.IP, "translate.google.com")
		if askForConfirmation() {
			writeToHosts(fastestRecord.IP)
		}
	}
}

func askForConfirmation() bool {
	var confirm string
	fmt.Println("Whether to write to the hosts file (yes/no):")
	fmt.Scanln(&confirm)
	switch strings.ToLower(confirm) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		slog.Info("Please type (y)es or (n)o and then press enter:")
		return askForConfirmation()
	}
}

// writeToHosts: only use for Google Translate
func writeToHosts(ip string) {
	var hostsFile string
	switch runtime.GOOS {
	case "windows":
		hostsFile = "C:\\Windows\\System32\\drivers\\etc\\hosts"
	case "darwin":
		hostsFile = "/private/etc/hosts"
	case "linux":
		hostsFile = "/etc/hosts"
	default:
		slog.Info("Your operating system is unknown, please configure hosts yourself.")
		return
	}
	backupPath := "hosts"
	err := Copy(hostsFile, backupPath)
	if err != nil {
		slog.Error("Backup hosts failed, please modify the hosts file yourself.", err)
		return
	}
	err = modifyHosts(hostsFile, ip)
	if err != nil {
		slog.Error("Modify hosts failed, please modify the hosts file yourself.", err)
		return
	}
	slog.Info("Successfully written to hosts file")
}

func Copy(srcPath, dstPath string) (err error) {
	r, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	// ignore error: file was opened read-only.
	defer r.Close()
	w, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() {
		err := w.Close()
		if err != nil {
		}
	}()
	_, err = io.Copy(w, r)
	return err
}

// modifyHosts: only use for Google Translate
func modifyHosts(hostsFile string, ip string) error {
	f, err := os.OpenFile(hostsFile, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	lineSeparator := "\n"
	if runtime.GOOS == "windows" {
		lineSeparator = "\r\n"
	}
	var builder strings.Builder
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "translate.googleapis.com") ||
			strings.Contains(line, "translate.google.com") {
			continue
		}
		builder.WriteString(line + lineSeparator)
	}
	builder.WriteString("")
	builder.WriteString(fmt.Sprintf("%s\t%s", ip, "translate.googleapis.com") + lineSeparator)
	builder.WriteString(fmt.Sprintf("%s\t%s", ip, "translate.google.com") + lineSeparator)
	err = f.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = f.WriteString(builder.String())
	if err != nil {
		return err
	}
	return nil
}
