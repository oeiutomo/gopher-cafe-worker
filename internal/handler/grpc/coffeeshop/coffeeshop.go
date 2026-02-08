//go:generate go tool mockgen -source=$GOFILE -destination=mock_coffeeshop_test.go -package=$GOPACKAGE
package coffeeshop

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	entity "gopher-cafe/internal/entity/coffeeshop"

	"github.com/ajaibid/coin-common-golang/logger"
	pb "github.com/rexyajaib/gopher-cafe/pkg/gen/go/v1"
)

type CoffeeshopUsecase interface {
	ExecuteBrew(ctx context.Context, orders []entity.Order, baristas int) []entity.OrderResult
	GetStats() (int64, int64, int64)
}

// Handler implements the gophercafepb.GopherCafeServiceServer interface
type CoffeeshopGrpcHandler struct {
	pb.UnimplementedGopherCafeServiceServer
	uc CoffeeshopUsecase
}

func NewCoffeeshopGrpcHandler(uc CoffeeshopUsecase) *CoffeeshopGrpcHandler {
	return &CoffeeshopGrpcHandler{
		uc: uc,
	}
}

// ExecuteBrew (CRP-01) triggers the simulation
func (h *CoffeeshopGrpcHandler) ExecuteBrew(ctx context.Context, req *pb.ExecuteBrewRequest) (*pb.ExecuteBrewResponse, error) {
	// TODO: put 2s as config
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	logger.Infof("Incoming request: %+v", req)
	// 1. CRP-01: Validation
	if !(req.Baristas >= 1) {
		return nil, status.Error(codes.InvalidArgument, "at least 1 barista is required")
	}
	if len(req.Orders) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least 1 order is required")
	}

	// 2. Mapping: Protobuf -> Domain Entities (CRP-08)
	internalOrders := make([]entity.Order, len(req.Orders))
	for i, o := range req.Orders {
		if o.Id <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "invalid order id at index %d", i)
		}

		internalOrders[i] = entity.Order{
			ID:    o.Id,
			Drink: toEntityDrink(o.Drink),
		}
	}

	// 3. Execution: Call the Usecase
	resultChan := make(chan []entity.OrderResult, 1)
	go func() {
		resultChan <- h.uc.ExecuteBrew(ctx, internalOrders, int(req.Baristas))
	}()

	var results []entity.OrderResult
	select {
	case results = <-resultChan:
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "context timeout exceeded")
	}

	// 4. Mapping: Domain Entities -> Protobuf Response (CRP-05)
	protoResults := make([]*pb.Result, len(results))
	for i, res := range results {
		protoSteps := make([]*pb.Step, len(res.Steps))
		for j, step := range res.Steps {
			protoSteps[j] = &pb.Step{
				Equipment: toPbEquipment(step.Equipment),
				StartMs:   step.StartTimeMs,
				EndMs:     step.EndTimeMs,
			}
		}

		protoResults[i] = &pb.Result{
			OrderId: res.OrderID,
			Steps:   protoSteps,
		}
	}

	return &pb.ExecuteBrewResponse{
		Results: protoResults,
	}, nil
}

// GetStats (CRP-07) retrieves aggregated simulation statistics
func (h *CoffeeshopGrpcHandler) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	totalRequests, _, p90RequestsMs := h.uc.GetStats()

	return &pb.GetStatsResponse{
		TotalRequestProcessed:     totalRequests,
		P90ProcessingMilliseconds: p90RequestsMs,
	}, nil
}
