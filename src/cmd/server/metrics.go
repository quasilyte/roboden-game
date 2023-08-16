package main

import (
	"sync/atomic"
)

type serverMetrics struct {
	data metricsData
}

type metricsData struct {
	Uptime float64 // In seconds

	MetricsSeq int

	// Request counters.
	NumReqErrors       int64
	ReqGetPlayerBoard  int64
	ReqGetBoard        int64
	ReqSavePlayerScore int64
	ReqVersion         int64

	NumReplaysQueued    int64
	NumReplaysCompleted int64
	NumReplaysFailed    int64
	NumReplaysRejected  int64
}

func (m *serverMetrics) IncNumReplaysQueued() {
	atomic.AddInt64(&m.data.NumReplaysQueued, 1)
}

func (m *serverMetrics) IncNumReplaysCompleted() {
	atomic.AddInt64(&m.data.NumReplaysCompleted, 1)
}

func (m *serverMetrics) IncNumReplaysFailed() {
	atomic.AddInt64(&m.data.NumReplaysFailed, 1)
}

func (m *serverMetrics) IncNumReplaysRejected() {
	atomic.AddInt64(&m.data.NumReplaysRejected, 1)
}

func (m *serverMetrics) IncNumReqErrors() {
	atomic.AddInt64(&m.data.NumReqErrors, 1)
}

func (m *serverMetrics) IncReqGetPlayerBoard() {
	atomic.AddInt64(&m.data.ReqGetPlayerBoard, 1)
}

func (m *serverMetrics) IncReqGetBoard() {
	atomic.AddInt64(&m.data.ReqGetBoard, 1)
}

func (m *serverMetrics) IncReqSavePlayerScore() {
	atomic.AddInt64(&m.data.ReqSavePlayerScore, 1)
}

func (m *serverMetrics) IncReqVersion() {
	atomic.AddInt64(&m.data.ReqVersion, 1)
}
