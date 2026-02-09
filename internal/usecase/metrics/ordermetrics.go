package coffeeshop

import (
	"github.com/rcrowley/go-metrics"
	"gopher-cafe/internal/entity/coffeeshop"
	"sync"
	"sync/atomic"
)

type OrderMetrics struct {
	totalRequests int64
	totalOrders   int64

	histogram metrics.Histogram

	mu sync.Mutex
}

func NewOrderMetrics(maxDurations uint32) *OrderMetrics {
	return &OrderMetrics{
		histogram: metrics.NewHistogram(metrics.NewExpDecaySample(1028, 0.015)),
	}
}

func (m *OrderMetrics) RecordTotalRequests(ordersSize int) {
	atomic.AddInt64(&m.totalRequests, int64(ordersSize))
}

func (m *OrderMetrics) RecordOrder(res coffeeshop.OrderResult) {
	atomic.AddInt64(&m.totalOrders, 1)

	if len(res.Steps) == 0 {
		return
	}

	duration := res.Steps[len(res.Steps)-1].EndTimeMs -
		res.Steps[0].StartTimeMs

	m.histogram.Update(duration)
}

func (m *OrderMetrics) GetStats() (int64, int64, int64) {
	totalReq := atomic.LoadInt64(&m.totalRequests)

	snap := m.histogram.Snapshot()

	return totalReq, snap.Count(), int64(snap.Percentile(0.9))
}
