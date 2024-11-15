package common

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/netip"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	maxWorkers = 10
	batchSize  = 1000
	// IPv6采样数量
	ipv6SampleSize = 1000
)

// IPProcessor 提供IP处理功能
type IPProcessor struct {
	resultChan chan string
	errorChan  chan error
	done       chan struct{}
}

// NewIPProcessor 创建新的处理器
func NewIPProcessor() *IPProcessor {
	return &IPProcessor{
		resultChan: make(chan string, batchSize),
		errorChan:  make(chan error, maxWorkers),
		done:       make(chan struct{}),
	}
}

// generateRandomIPv6 在给定前缀内生成随机IPv6地址
func generateRandomIPv6(prefix netip.Prefix) (netip.Addr, error) {
	// 获取网络前缀部分
	network := prefix.Addr().AsSlice()
	ones := prefix.Bits()

	// 计算需要随机化的字节数
	randomBytes := 16 - (ones+7)/8

	// 复制原始地址
	result := make([]byte, 16)
	copy(result, network)

	// 随机化剩余字节
	for i := 16 - randomBytes; i < 16; i++ {
		result[i] = byte(rand.Intn(256))
	}

	// 转换回netip.Addr
	addr, ok := netip.AddrFromSlice(result)
	if !ok {
		return netip.Addr{}, fmt.Errorf("failed to create IPv6 address")
	}

	return addr, nil
}

// processCIDR 处理单个CIDR区间
func processCIDR(cidr string, resultChan chan<- string) error {
	p, err := netip.ParsePrefix(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR %s: %w", cidr, err)
	}

	p = p.Masked()

	if p.Addr().Is6() {
		// 对IPv6进行随机采样
		slog.Info("Processing IPv6 CIDR with random sampling", "range", cidr)
		for i := 0; i < ipv6SampleSize; i++ {
			addr, err := generateRandomIPv6(p)
			if err != nil {
				slog.Error("Failed to generate IPv6 address", "error", err)
				continue
			}

			select {
			case resultChan <- addr.String():
			case <-time.After(100 * time.Millisecond):
				slog.Warn("Failed to send IPv6", "ip", addr.String())
			}
		}
	} else {
		// IPv4处理逻辑保持不变
		addr := p.Addr()
		for {
			if !p.Contains(addr) {
				break
			}

			select {
			case resultChan <- addr.String():
			case <-time.After(100 * time.Millisecond):
				slog.Warn("Failed to send IPv4", "ip", addr.String())
			}

			addr = addr.Next()
		}
	}

	return nil
}

// GetIPs 主函数
func GetIPs(config *Config) []string {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())

	cidrs, err := loadCIDRs(config)
	if err != nil {
		slog.Error("Failed to load CIDRs", "error", err)
		return nil
	}

	processor := NewIPProcessor()
	cidrChan := make(chan string, maxWorkers)
	var wg sync.WaitGroup

	// 启动工作协程
	workerCount := min(maxWorkers, len(cidrs))
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for cidr := range cidrChan {
				if err := processCIDR(cidr, processor.resultChan); err != nil {
					select {
					case processor.errorChan <- err:
					default:
						slog.Error("Process CIDR failed", "cidr", cidr, "error", err)
					}
				}
			}
		}()
	}

	// 收集结果
	var results []string
	resultBatch := make([]string, 0, batchSize)

	go func() {
		defer close(processor.done)
		for ip := range processor.resultChan {
			resultBatch = append(resultBatch, ip)

			if len(resultBatch) >= batchSize {
				newBatch := make([]string, len(resultBatch))
				copy(newBatch, resultBatch)
				results = append(results, newBatch...)
				resultBatch = resultBatch[:0]
			}
		}
		if len(resultBatch) > 0 {
			results = append(results, resultBatch...)
		}
	}()

	// 发送CIDR到workers
	for _, cidr := range cidrs {
		select {
		case cidrChan <- cidr:
		case err := <-processor.errorChan:
			slog.Error("CIDR processing error", "error", err)
		}
	}
	close(cidrChan)

	wg.Wait()
	close(processor.resultChan)
	<-processor.done

	slog.Info("IP加载完成", "数量", len(results))
	return results
}

// isIPv4 检查是否为IPv4地址
func isIPv4(ip string) bool {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}
	return addr.Is4()
}

func loadCIDRs(config *Config) ([]string, error) {
	siteCfg := RetrieveSiteCfg(config)
	targetFile := siteCfg.CustomIPRangesFile
	withIPv6 := siteCfg.WithIPv6

	// 检查自定义文件是否存在
	if _, err := os.Stat(targetFile); err != nil {
		if os.IsNotExist(err) {
			slog.Warn("Custom IP ranges file not found, using default")
			targetFile = siteCfg.IPRangesFile
		} else {
			return nil, fmt.Errorf("failed to check IP ranges file: %w", err)
		}
	}

	f, err := os.Open(targetFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open IP ranges file %s: %w", targetFile, err)
	}
	defer f.Close()

	var cidrs []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// 跳过空行和注释行
		if line == "" || (line[0] == '#') {
			continue
		}
		// 如果不支持IPv6且当前行是IPv6地址，则跳过
		if !withIPv6 && !isIPv4(line) {
			continue
		}
		cidrs = append(cidrs, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read IP ranges file: %w", err)
	}
	slog.Info("Load CIDRs:", "Count", len(cidrs))
	return cidrs, nil
}

// RetrieveSiteCfg 获取站点配置
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

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// writeToFile 写入文件
func writeToFile(scanRecords ScanRecordArray, config *Config) {
	siteCfg := RetrieveSiteCfg(config)
	outputFile := siteCfg.IPOutputFile
	f, err := os.Create(outputFile)
	if err != nil {
		slog.Error("Failed to create file", "error", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, record := range scanRecords {
		if _, err := w.WriteString(record.IP + "\n"); err != nil {
			slog.Error("Write to output file failed", "error", err)
		}
	}

	if err := w.Flush(); err != nil {
		slog.Error("Flush failed", "error", err)
	}
}

// writeToHosts 写入hosts文件
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
		slog.Info("Unknown operating system, please configure hosts manually")
		return
	}

	if err := backupHosts(hostsFile); err != nil {
		slog.Error("Backup hosts failed", "error", err)
		return
	}

	if err := modifyHosts(hostsFile, ip, domains); err != nil {
		slog.Error("Modify hosts failed", "error", err)
		return
	}

	slog.Info("Successfully written to hosts file")
}

// backupHosts 备份hosts文件
func backupHosts(hostsFile string) error {
	return Copy(hostsFile, "hosts.backup")
}

// Copy 文件复制
func Copy(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// modifyHosts 修改hosts文件
func modifyHosts(hostsFile string, ip string, domains []string) error {
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	gLabel := "# Google Translate"

	// 删除旧的条目
	for _, line := range lines {
		if !strContainSlice(line, domains) && !strings.Contains(line, gLabel) {
			newLines = append(newLines, line)
		}
	}

	// 添加新的条目
	newLines = append(newLines, "", gLabel)
	for _, domain := range domains {
		newLines = append(newLines, fmt.Sprintf("%s %s", ip, domain))
	}

	// 写入文件
	return os.WriteFile(hostsFile, []byte(strings.Join(newLines, "\n")), 0644)
}

func printResult(scanRecords ScanRecordArray, config *Config) {
	if len(scanRecords) == 0 {
		site := RetrieveSiteCfg(config)
		customIPRangesFile := site.CustomIPRangesFile
		_, err := os.Stat(customIPRangesFile)
		if err == nil {
			slog.Error("No available ip found! Please delete the %s file and re-run will scan all IP ranges", customIPRangesFile)
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
	for _, domain := range siteCfg.Domains {
		fmt.Printf("%v\t%s\n", fastestRecord.IP, domain)
	}
	if askForConfirmation() {
		writeToHosts(fastestRecord.IP, siteCfg.Domains)
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

// strContainSlice 检查字符串是否包含切片中的任何元素
func strContainSlice(s string, ls []string) bool {
	for _, substr := range ls {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
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
