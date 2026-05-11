package grpc

import (
	"context"

	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
)

type HealthHandler struct {
	xov1.UnimplementedHealthServiceServer
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(ctx context.Context, req *xov1.HealthRequest) (*xov1.HealthResponse, error) {
	return &xov1.HealthResponse{
		Status:  xov1.HealthStatus_HEALTH_STATUS_SERVING,
		Message: "ok",
	}, nil
}
