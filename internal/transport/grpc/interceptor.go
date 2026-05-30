package grpc

import (
	"context"
	"log/slog"
	"time"

	"github.com/MehrshadFb/xo-grpc/internal/metrics"
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

	metrics.GRPCRequestsTotal.WithLabelValues(
		info.FullMethod,
		code.String(),
		"unary",
	).Inc()

	metrics.GRPCRequestDurationSeconds.WithLabelValues(
		info.FullMethod,
		code.String(),
		"unary",
	).Observe(duration.Seconds())

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

func LoggingStreamInterceptor(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()

	err := handler(srv, ss)

	duration := time.Since(start)
	code := status.Code(err)

	metrics.GRPCRequestsTotal.WithLabelValues(
		info.FullMethod,
		code.String(),
		"stream",
	).Inc()

	metrics.GRPCRequestDurationSeconds.WithLabelValues(
		info.FullMethod,
		code.String(),
		"stream",
	).Observe(duration.Seconds())

	if err != nil {
		slog.Error(
			"grpc stream request failed",
			"method", info.FullMethod,
			"code", code.String(),
			"duration", duration.String(),
			"client_stream", info.IsClientStream,
			"server_stream", info.IsServerStream,
			"error", err,
		)
	} else {
		slog.Info(
			"grpc stream request completed",
			"method", info.FullMethod,
			"code", code.String(),
			"duration", duration.String(),
			"client_stream", info.IsClientStream,
			"server_stream", info.IsServerStream,
		)
	}

	return err
}
