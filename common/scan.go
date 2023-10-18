package common

import (
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"
)

func testOneDetail(destination string, destinationPort uint16, config *Config, record *ScanRecord) bool {
	record.IP = destination
	record.Protocol = config.Ping.Protocol
	if config.Ping.Protocol == "udp" || config.Ping.Protocol == "tcp" {
		record.IP += fmt.Sprintf(":%d", destinationPort)
	}
	slog.Info("Start Ping:", "IP", record.IP)
	successTimes := 0
	var latencies []float64
	for i := 0; i < config.Ping.Count; i += 1 {
		var err error
		// startTime for calculating the latency/RTT
		startTime := time.Now()

		switch config.Ping.Protocol {
		case "icmp":
			err = pingIcmp(destination, config.Ping.Timeout)
		case "tcp":
			err = pingTcp(destination, destinationPort, config.Ping.Timeout)
		case "udp":
			err = pingUdp(destination, destinationPort, config.Ping.Timeout)
		}
		//store the time elapsed before processing potential errors
		latency := time.Since(startTime).Seconds() * 1000

		// evaluate potential ping failures
		if err != nil {
			switch err.Error() {
			case ErrorTimeout:
				latency = 9999999
			case ErrorConnRefused:
				latency = -1
			default:
				latency = 0
			}
		}
		switch latency {
		case -1, 0:
			// do nothing
		case 9999999:
			// For udp, a timeout indicates that the port *maybe* open.
			if config.Ping.Protocol == "udp" {
				successTimes += 1
			}
		default:
			successTimes += 1
		}
		latencies = append(latencies, latency)
		// sleep 20 milliseconds between pings to prevent floods
		time.Sleep(100 * time.Millisecond)
	}
	sum := 0.0
	for i := 0; i < len(latencies); i++ {
		sum += latencies[i]
	}
	record.Latency = math.Round(sum / float64(len(latencies)))
	success := false
	if successTimes == config.Ping.Count {
		success = true
	}
	return success
}

func testOne(ch chan string, config *Config, scanResult *ScanResult, wg *sync.WaitGroup) {
	for destination := range ch {
		if destination == "" {
			slog.Info("============== waitgroup done =================")
			wg.Done()
			break
		}
		record := new(ScanRecord)
		destinationPort := config.Ping.Port
		success := testOneDetail(destination, destinationPort, config, record)
		if success {
			success = HttpPing(destination, destinationPort, config)
			if success {
				scanResult.AddRecord(record)
			}
			scanResult.IncScanCounter()
		}
	}
}

func Start(config *Config) {
	scanResult := new(ScanResult)
	ips := GetIPs(config)
	workers := config.General.Workers
	ch := make(chan string, len(ips))
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go testOne(ch, config, scanResult, &wg)
	}
	for _, destination := range ips {
		ch <- destination
	}
	for i := 0; i < workers; i++ {
		ch <- ""
	}
	wg.Wait()
	close(ch)
	scanRecords := scanResult.scanRecords
	sort.Slice(scanRecords, func(i, j int) bool {
		return scanRecords[i].Latency < scanRecords[j].Latency
	})
	writeToFile(scanRecords, config)
	printResult(scanRecords)
}
