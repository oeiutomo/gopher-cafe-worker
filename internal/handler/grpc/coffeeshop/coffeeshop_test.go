package coffeeshop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	entity "gopher-cafe/internal/entity/coffeeshop"

	pb "github.com/rexyajaib/gopher-cafe/pkg/gen/go/v1"
)

func TestExecuteBrew(t *testing.T) {
	// Initialize the gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create the generated mock
	mockUC := NewMockCoffeeshopUsecase(ctrl)
	handler := NewCoffeeshopGrpcHandler(nil, mockUC)

	// Define test cases
	tests := []struct {
		name         string
		req          *pb.ExecuteBrewRequest
		mockExpect   func()
		expectedCode codes.Code
		expectedRes  bool // true if we expect a non-nil response
	}{
		{
			name: "Success - Single Order",
			req: &pb.ExecuteBrewRequest{
				Baristas: 1,
				Orders: []*pb.Order{
					{Id: 101, Drink: pb.DrinkType_DRINK_TYPE_ESPRESSO},
				},
			},
			mockExpect: func() {
				// We expect the usecase to be called exactly once
				mockUC.EXPECT().
					ExecuteBrew(context.Background(), gomock.Len(1), 1).
					Return([]entity.OrderResult{
						{
							OrderID: 101,
							Steps: []entity.StepExecution{
								{Equipment: entity.EquipGrinder, StartTimeMs: 10, EndTimeMs: 15},
							},
						},
					})
			},
			expectedCode: codes.OK,
			expectedRes:  true,
		},
		{
			name: "Error - Invalid Baristas (CRP-01)",
			req: &pb.ExecuteBrewRequest{
				Baristas: 0,
				Orders:   []*pb.Order{{Id: 1, Drink: pb.DrinkType_DRINK_TYPE_ESPRESSO}},
			},
			mockExpect:   func() {}, // Usecase should NOT be called
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error - Empty Orders (CRP-01)",
			req: &pb.ExecuteBrewRequest{
				Baristas: 1,
				Orders:   []*pb.Order{},
			},
			mockExpect:   func() {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error - Invalid Order ID (CRP-01)",
			req: &pb.ExecuteBrewRequest{
				Baristas: 1,
				Orders:   []*pb.Order{{Id: 0, Drink: pb.DrinkType_DRINK_TYPE_ESPRESSO}},
			},
			mockExpect:   func() {},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup expectations for this specific test case
			tt.mockExpect()

			// Execute the call
			resp, err := handler.ExecuteBrew(context.Background(), tt.req)

			// Assertions
			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				if tt.expectedRes {
					assert.NotNil(t, resp)
					assert.Equal(t, tt.req.Orders[0].Id, resp.Results[0].OrderId)
				}
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}
