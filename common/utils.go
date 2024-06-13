package common

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/netip"
	"os"
	"reflect"
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

func loadCIDRs(config *Config) ([]string, error) {
	var cidrs []string
	siteCfg := RetrieveSiteCfg(config)
	customIPRangesFile := siteCfg.CustomIPRangesFile
	ipRangesFile := siteCfg.IPRangesFile
	withIPv6 := siteCfg.WithIPv6
	targetFile := customIPRangesFile
	_, err := os.Stat(targetFile)
	if err == nil {
		slog.Info("found custom ip ranges file.")
	} else if os.IsNotExist(err) {
		slog.Warn("custom ip ranges file does not exist, use default ip ranges file instead!")
		targetFile = ipRangesFile
	} else {
		slog.Warn("custom ip ranges file stat error: use default ip ranges file instead!", "error", err)
		targetFile = ipRangesFile
	}
	f, err := os.Open(targetFile)
	if err != nil {
		slog.Error("Could not open ip address ranges:", "file", targetFile)
		return cidrs, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !withIPv6 && !isIPv4(line) {
			continue
		}
		cidrs = append(cidrs, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		slog.Error("Could not load ip address ranges file:", targetFile)
		return cidrs, err
	}
	return cidrs, err
}

func GetIPs(config *Config) []string {
	var iparr IPArray
	cidrs, err := loadCIDRs(config)
	if err != nil {
		slog.Error("get ips failed!")
		return iparr.IPs
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

// RetrieveSiteCfg assumed that the site name must exist in the configuration file and no error handling required
func RetrieveSiteCfg(config *Config) Site {
	siteName := config.General.Site
	sv := reflect.ValueOf(config.Sites)
	sites := sv.Interface().([]Site)
	for _, site := range sites {
		if site.Name == siteName {
			return site
		}
	}
	return Site{}
}

func writeToFile(scanRecords ScanRecordArray, config *Config) {
	siteCfg := RetrieveSiteCfg(config)
	outputFile := siteCfg.IPOutputFile
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
		site := RetrieveSiteCfg(config)
		customIPRangesFile := site.CustomIPRangesFile
		_, err := os.Stat(customIPRangesFile)
		if err == nil {
			log.Printf("No available ip found! Please delete the %s file and re-run will scan all IP ranges", customIPRangesFile)
		} else {
			slog.Info("No available ip found!")
		}
		return
	}
	head := scanRecords
	if len(head) > 10 {
		head = head[:10]
	}
	fmt.Printf("%s\t%s\t%s\t%s\n", "IP", "Protocol", "PingRTT", "HttpRTT")
	for _, record := range head {
		fmt.Printf("%s\t%s\t%.f\t%.f\n", record.IP, record.Protocol, record.PingRTT, record.HttpRTT)
	}
	fastestRecord := *scanRecords[0]
	slog.Info("The fastest IP has been found:")
	siteCfg := RetrieveSiteCfg(config)
	if len(siteCfg.Domains) > 0 {
		for _, domain := range siteCfg.Domains {
			fmt.Printf("%v\t%s\n", fastestRecord.IP, domain)
		}
	
		if askForConfirmation() {
			writeToHosts(fastestRecord.IP, siteCfg.Domains)
		}
	} else {
		slog.Warn("No domains found in the configuration.")
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

func writeToHosts(ip string, domains []string) {
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
	err = modifyHosts(hostsFile, ip, domains)
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
func modifyHosts(hostsFile string, ip string, domains []string) error {
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
		if strContainSlice(line, domains) {
			continue
		}
		builder.WriteString(line + lineSeparator)
	}
	builder.WriteString(lineSeparator)
	for _, domain := range domains {
		line := fmt.Sprintf("%s\t%s", ip, domain) + lineSeparator
		builder.WriteString(line)
	}
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

func AssertSiteName(config *Config) bool {
	siteName := config.General.Site
	for _, site := range config.Sites {
		if site.Name == siteName {
			return true
		}
	}
	return false
}

func strContainSlice(s string, ls []string) bool {
	for _, substr := range ls {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
