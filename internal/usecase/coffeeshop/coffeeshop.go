package coffeeshop

import (
	"context"
	"github.com/ajaibid/coin-common-golang/logger"
	coffeeshop "gopher-cafe/internal/usecase/metrics"
	"sync"
	"time"

	entity "gopher-cafe/internal/entity/coffeeshop"
)

type CoffeeshopUsecase struct {
	orderMetrics *coffeeshop.OrderMetrics
}

func NewCoffeeshopUsecase(ordMetrics *coffeeshop.OrderMetrics) *CoffeeshopUsecase {
	return &CoffeeshopUsecase{
		orderMetrics: ordMetrics,
	}
}

func (u *CoffeeshopUsecase) ExecuteBrew(ctx context.Context, orders []entity.Order, baristas int) []entity.OrderResult {
	// add number of orders as global totalRequests
	u.orderMetrics.RecordTotalRequests(len(orders))

	results := make([]entity.OrderResult, 0, len(orders))

	// order channel as goroutine input
	orderChan := make(chan entity.Order, len(orders))
	for _, ord := range orders {
		orderChan <- ord
	}
	close(orderChan)

	// result channel as goroutine output
	resultChan := make(chan entity.OrderResult, len(orders))

	// WG for barista goroutines
	var wg sync.WaitGroup
	wg.Add(baristas)

	for i := 0; i < baristas; i++ {
		// barista goroutine
		go func(id int) {
			defer wg.Done()

			// 1 barista take 1 order from order channel
			for order := range orderChan {
				recipe := entity.Recipes[order.Drink]

				res := entity.OrderResult{OrderID: order.ID}
				for _, step := range recipe {
					startStep := time.Now().UnixMilli()

					// processOrder
					err := u.processStep(ctx, step)
					if err != nil {
						logger.ErrorKV("failed to processStep",
							logger.KV("order", order), logger.KV("step", step), logger.KV("error", err))
						break
					}

					res.Steps = append(res.Steps, entity.StepExecution{
						Equipment:   step.Equipment,
						StartTimeMs: startStep,
						EndTimeMs:   time.Now().UnixMilli(),
					})
				}

				u.orderMetrics.RecordOrder(res)
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

func (u *CoffeeshopUsecase) processStep(ctx context.Context, step *entity.RecipeStep) error {
	select {
	case step.Semaphore <- struct{}{}:
		// release semaphore
		defer func() { <-step.Semaphore }()

		return u.doProcessStep(ctx, step)

	case <-ctx.Done():
		logger.Warn("context TIMEOUT on waiting for semaphore")
		return ctx.Err()
	}
}

func (u *CoffeeshopUsecase) doProcessStep(ctx context.Context, step *entity.RecipeStep) error {
	timer := time.NewTimer(step.Duration)
	// stop timer
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil

	case <-ctx.Done():
		logger.Warn("context TIMEOUT on during processing step")
		return ctx.Err()
	}
}

func (u *CoffeeshopUsecase) GetStats() (int64, int64, int64) {
	return u.orderMetrics.GetStats()
}
