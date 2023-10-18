package common

import (
	"log/slog"
	"sync"
	"sync/atomic"
)

type ScanRecord struct {
	IP       string  `json:"ip"`        // ip:port
	Protocol string  `json:"protocol"`  // icmp, tcp, udp
	Latency  float64 `json:"latencies"` // response latency in milliseconds: 9999999 indicates timeout, -1 indicates unreachable, 0 general error.
}

type ScanRecordArray []*ScanRecord

type ScanResult struct {
	scanned     int32
	scanRecords ScanRecordArray
	recordMutex sync.Mutex
	hostsMutex  sync.Mutex
}

func (records *ScanRecordArray) Len() int {
	return len(*records)
}

func (records *ScanRecordArray) Less(i, j int) bool {
	return (*records)[i].Latency < (*records)[j].Latency
}

func (records *ScanRecordArray) Swap(i, j int) {
	tmp := (*records)[i]
	(*records)[i] = (*records)[j]
	(*records)[j] = tmp
}

func (result *ScanResult) AddRecord(record *ScanRecord) {
	result.recordMutex.Lock()
	if result.scanRecords == nil {
		result.scanRecords = make(ScanRecordArray, 0)
	}
	result.scanRecords = append(result.scanRecords, record)
	result.recordMutex.Unlock()
	slog.Info("Found an IP:", slog.String("IP", record.IP), slog.Float64("Latency", record.Latency))
}

func (result *ScanResult) IncScanCounter() {
	atomic.AddInt32(&(result.scanned), 1)
	if result.scanned%1000 == 0 {
		slog.Info("Scanned:", result.scanned)
	}
}
