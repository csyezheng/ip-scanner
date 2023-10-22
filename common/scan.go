package common

import (
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"
)

func pingOneIP(destination string, destinationPort uint16, config *Config, record *ScanRecord) bool {
	slog.Debug("Start Ping:", "IP", destination)
	record.IP = destination
	record.Protocol = config.Ping.Protocol
	if config.Ping.Protocol == "udp" || config.Ping.Protocol == "tcp" {
		record.IP += fmt.Sprintf(":%d", destinationPort)
	}
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
	record.PingRTT = math.Round(sum / float64(len(latencies)))
	success := false
	if (config.Ping.all && successTimes == config.Ping.Count) || (!config.Ping.all && successTimes > 0) {
		success = true
	}
	return success
}

func reqOneIP(destination string, destinationPort uint16, config *Config, record *ScanRecord) bool {
	slog.Debug("Start Ping:", "IP", destination)
	successTimes := 0
	var latencies []float64
	for i := 0; i < config.HTTP.Count; i += 1 {
		var err error
		// startTime for calculating the latency/RTT
		startTime := time.Now()

		err = reqHEAD(destination, destinationPort, config)
		//store the time elapsed before processing potential errors
		latency := time.Since(startTime).Seconds() * 1000

		// evaluate potential ping failures
		if err != nil {
			latency = 9999999
		} else {
			successTimes += 1
		}
		latencies = append(latencies, latency)
		// sleep 20 milliseconds between request to prevent floods
		time.Sleep(100 * time.Millisecond)
	}
	sum := 0.0
	for i := 0; i < len(latencies); i++ {
		sum += latencies[i]
	}
	record.HttpRTT = math.Round(sum / float64(len(latencies)))
	success := false
	if (config.HTTP.all && successTimes == config.HTTP.Count) || (!config.HTTP.all && successTimes > 0) {
		success = true
	}
	return success
}

func testOne(ch chan string, config *Config, scanResult *ScanResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for destination := range ch {
		if (config.General.ScannedLimit > 0 && config.General.ScannedLimit < scanResult.Scanned()) ||
			(config.General.FoundLimit > 0 && config.General.FoundLimit < scanResult.Found()) {
			slog.Debug("The limit number of scans from configuration file has been reached, stop scanning!")
			return
		}
		record := new(ScanRecord)
		destinationPort := config.Ping.Port
		success := pingOneIP(destination, destinationPort, config, record)
		if success {
			success = reqOneIP(destination, destinationPort, config, record)
			if success {
				scanResult.AddRecord(record)
			} else {
				slog.Debug(fmt.Sprintf("IP %s http test timeout", destination))
			}
			scanResult.IncScanCounter()
		} else {
			slog.Debug(fmt.Sprintf("IP %s ping test timeout", destination))
		}
	}
}

func Run(config *Config) {
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
	// Sender close a channel to indicate that no more values will be sent.
	close(ch)
	wg.Wait()
	scanRecords := scanResult.scanRecords
	sort.Slice(scanRecords, func(i, j int) bool {
		return scanRecords[i].HttpRTT < scanRecords[j].HttpRTT
	})
	writeToFile(scanRecords, config)
	printResult(scanRecords, config)
}
