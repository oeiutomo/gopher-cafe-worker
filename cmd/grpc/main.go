package main

import (
	coffeeshop "gopher-cafe/internal/usecase/metrics"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	appCfg "gopher-cafe/config"
	handler "gopher-cafe/internal/handler/grpc/coffeeshop"
	usecase "gopher-cafe/internal/usecase/coffeeshop"

	pb "github.com/rexyajaib/gopher-cafe/pkg/gen/go/v1"

	"github.com/ajaibid/coin-common-golang/config"
	"github.com/ajaibid/coin-common-golang/logger"
)

func main() {
	// Create a TCP Listener on a specific port
	var cfg appCfg.Config
	err := config.LoadConfig(&cfg, "config/.env")
	if err != nil {
		log.Fatal(err)
	}
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Grpc.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	logger.NewLogger(logger.LoggerConf{
		LogLevel:     cfg.Logger.LogLevel,
		LogFormatter: cfg.Logger.LogFormatter,
	})

	// Initialize the Layers
	orderMetrics := coffeeshop.NewOrderMetrics(cfg.Cafe.OrderDurationsCap)
	coffeeUsecase := usecase.NewCoffeeshopUsecase(orderMetrics)
	coffeeHandler := handler.NewCoffeeshopGrpcHandler(&cfg.Cafe, coffeeUsecase)

	// Create the gRPC Server instance
	grpcServer := grpc.NewServer()

	// Register the Service (The "Route Definition")
	// This tells the gRPC server to route incoming GopherCafe calls to our handler.
	pb.RegisterGopherCafeServiceServer(grpcServer, coffeeHandler)

	// Optional: Enable reflection.
	// This allows tools like Postman or 'evans' to "see" your endpoints automatically.
	reflection.Register(grpcServer)

	// Start Serving
	log.Printf("Coffee Shop Simulation Server is running on %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
