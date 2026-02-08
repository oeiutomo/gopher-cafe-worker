package coffeeshop

import (
	"context"
	"github.com/ajaibid/coin-common-golang/logger"
	"sync"
	"sync/atomic"
	"time"

	entity "gopher-cafe/internal/entity/coffeeshop"
)

type CoffeeshopUsecase struct {
	totalRequests int64
	totalOrders   int64
	p90RequestsMs int64
}

func NewCoffeeshopUsecase() *CoffeeshopUsecase {
	return &CoffeeshopUsecase{}
}

func (u *CoffeeshopUsecase) ExecuteBrew(ctx context.Context, orders []entity.Order, baristas int) []entity.OrderResult {
	// add number of orders as global totalRequests
	atomic.AddInt64(&u.totalRequests, int64(len(orders)))

	results := make([]entity.OrderResult, 0, len(orders))

	// order channel as goroutine input
	orderChan := make(chan entity.Order, len(orders))
	for _, ord := range orders {
		orderChan <- ord
	}
	close(orderChan)

	// result channel as goroutine output
	resultChan := make(chan entity.OrderResult, len(orders))

	// WG for barista, will indicates
	var wg sync.WaitGroup
	wg.Add(baristas)

	for i := 0; i < baristas; i++ {
		// barisa goroutine
		go func(id int) {
			defer wg.Done()

			// 1 barista take 1 order from order channel
			for order := range orderChan {
				recipe := entity.Recipes[order.Drink]

				res := entity.OrderResult{OrderID: order.ID}
				for _, step := range recipe {
					startStep := time.Now().UnixMilli()

					// processOrder
					if err := u.processStep(ctx, step.Duration); err != nil {
						return
					}

					res.Steps = append(res.Steps, entity.StepExecution{
						Equipment:   step.Equipment,
						StartTimeMs: startStep,
						EndTimeMs:   time.Now().UnixMilli(),
					})
				}

				u.recordOrderStats(res)
				resultChan <- res
			}
		}(i)
	}

	// wait until all baristas process all orders
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// convert result from channel to array
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

func (u *CoffeeshopUsecase) processStep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil

	case <-ctx.Done():
		logger.Warn("context TIMEOUT exceeded")
		return ctx.Err()
	}
}

func (u *CoffeeshopUsecase) recordOrderStats(res entity.OrderResult) {
	atomic.AddInt64(&u.totalOrders, 1)
	if len(res.Steps) > 0 {
		duration := res.Steps[len(res.Steps)-1].EndTimeMs - res.Steps[0].StartTimeMs
		atomic.AddInt64(&u.p90RequestsMs, duration)
	}
}

func (u *CoffeeshopUsecase) GetStats() (int64, int64, int64) {
	return atomic.LoadInt64(&u.totalRequests),
		atomic.LoadInt64(&u.totalOrders),
		atomic.LoadInt64(&u.p90RequestsMs)
}
