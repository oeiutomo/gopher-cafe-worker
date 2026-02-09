package interceptor

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

func TimeoutInterceptor(brewTimeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		timeout, cancelFunc := context.WithTimeout(ctx, brewTimeout)
		defer cancelFunc()
		return handler(timeout, req)
	}
}
