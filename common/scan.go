package common

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

type TestRecord struct {
	ResolvedHost string  `json:"resolvedhost"` // Information about the ip:port being pinged
	Protocol     string  `json:"protocol"`     // icmp, tcp, udp
	Latency      float64 `json:"latencies"`    // response latency in milliseconds: 9999999 indicates timeout, -1 indicates unreachable, 0 general error.
	Success      bool    `json:"Success"`      // ping success or failed
}

type TestRecordArray []TestRecord

func testOneDetail(destination string, destinationPort float64, config Config, record *TestRecord) bool {
	record.ResolvedHost = destination
	record.Protocol = config.Protocol
	if config.Protocol == "udp" || config.Protocol == "tcp" {
		record.ResolvedHost += fmt.Sprintf(":%.0f", destinationPort)
	}
	log.Printf("Start Ping: %s", record.ResolvedHost)
	successTimes := 0
	var latencies []float64
	for i := 0; i < config.Count; i += 1 {
		var err error
		// startTime for calculating the latency/RTT
		startTime := time.Now()

		switch config.Protocol {
		case "icmp":
			err = pingIcmp(destination, config.Timeout)
		case "tcp":
			err = pingTcp(destination, destinationPort, config.Timeout)
		case "udp":
			err = pingUdp(destination, destinationPort, config.Timeout)
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
			if config.Protocol == "udp" {
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
	if successTimes == config.Count {
		record.Success = true
	}
	return record.Success
}

func testOne(destination string, config Config, records *TestRecordArray, wg *sync.WaitGroup) {
	defer wg.Done()
	record := new(TestRecord)
	destinationPort := config.Port
	success := testOneDetail(destination, destinationPort, config, record)
	if !success {
		return
	}
	*records = append(*records, *record)
}

func BatchTest(config Config) TestRecordArray {
	records := new(TestRecordArray)
	ips := GetIps(config)
	log.Println(ips)
	var wg sync.WaitGroup
	for _, ip := range ips {
		wg.Add(1)
		go testOne(ip, config, records, &wg)
	}
	wg.Wait()
	return *records
}
