package common

import (
	"log/slog"
	"sync"
	"sync/atomic"
)

type ScanRecord struct {
	IP       string  `json:"ip"`       // ip:port
	Protocol string  `json:"protocol"` // icmp, tcp, udp
	PingRTT  float64 `json:"pingrtt"`  // response latency in milliseconds: 9999999 indicates timeout, -1 indicates unreachable, 0 general error.
	HttpRTT  float64 `json:"httprtt"`  // response latency in milliseconds: 9999999 indicates timeout, -1 indicates unreachable, 0 general error.
}

type ScanRecordArray []*ScanRecord

type ScanResult struct {
	scanned     int32
	scanRecords ScanRecordArray
	recordMutex sync.Mutex
}

func (records *ScanRecordArray) Len() int {
	return len(*records)
}

func (records *ScanRecordArray) Less(i, j int) bool {
	return (*records)[i].HttpRTT < (*records)[j].HttpRTT
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
	slog.Info("Found an IP:", slog.String("IP", record.IP), slog.Float64("PingRTT", record.PingRTT),
		slog.Float64("HttpRTT", record.HttpRTT))
}

func (result *ScanResult) IncScanCounter() {
	atomic.AddInt32(&(result.scanned), 1)
	if result.scanned%1000 == 0 {
		slog.Info("Progress:", "Scanned", result.scanned)
	}
}
