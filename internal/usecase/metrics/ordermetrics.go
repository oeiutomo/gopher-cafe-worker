package coffeeshop

import (
	"sort"
	"sync"
	"sync/atomic"

	entity "gopher-cafe/internal/entity/coffeeshop"
)

type OrderMetrics struct {
	totalRequests int64
	totalOrders   int64

	durations    []int64
	maxDurations uint32
	writeIdx     uint64

	mu sync.Mutex
}

func NewOrderMetrics(maxDurations uint32) *OrderMetrics {
	return &OrderMetrics{
		maxDurations: maxDurations,
		durations:    make([]int64, maxDurations),
	}
}

func (m *OrderMetrics) RecordTotalRequests(ordersSize int) {
	atomic.AddInt64(&m.totalRequests, int64(ordersSize))
}

func (m *OrderMetrics) RecordOrder(res entity.OrderResult) {
	atomic.AddInt64(&m.totalOrders, 1)

	if len(res.Steps) == 0 {
		return
	}

	duration := res.Steps[len(res.Steps)-1].EndTimeMs -
		res.Steps[0].StartTimeMs

	idx := atomic.AddUint64(&m.writeIdx, 1) - 1
	pos := idx % uint64(m.maxDurations)

	atomic.StoreInt64(&m.durations[pos], duration)
}

func (m *OrderMetrics) GetTotalRequests() int64 {
	return atomic.LoadInt64(&m.totalRequests)
}

func (m *OrderMetrics) GetTotalOrders() int64 {
	return atomic.LoadInt64(&m.totalOrders)
}

func (m *OrderMetrics) GetP90Duration() int64 {
	written := atomic.LoadUint64(&m.writeIdx)
	if written == 0 {
		return 0
	}

	count := uint32(written)
	if count > m.maxDurations {
		count = m.maxDurations
	}

	snapshot := make([]int64, count)
	for i := uint32(0); i < count; i++ {
		snapshot[i] = atomic.LoadInt64(&m.durations[i])
	}

	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i] < snapshot[j]
	})

	idx := int(float64(len(snapshot)) * 0.9)
	if idx >= len(snapshot) {
		idx = len(snapshot) - 1
	}

	return snapshot[idx]
}
