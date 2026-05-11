package grpc

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggingUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	code := status.Code(err)

	if err != nil {
		slog.Error(
			"grpc unary request failed",
			"method", info.FullMethod,
			"code", code.String(),
			"duration", duration.String(),
			"error", err,
		)
	} else {
		slog.Info(
			"grpc unary request completed",
			"method", info.FullMethod,
			"code", code.String(),
			"duration", duration.String(),
		)
	}

	return resp, err
}
