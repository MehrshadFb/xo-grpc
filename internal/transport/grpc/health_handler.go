package grpc

import (
	"context"

	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
	healthsvc "github.com/MehrshadFb/xo-grpc/internal/service/health"
)

type HealthHandler struct {
	xov1.UnimplementedHealthServiceServer

	service *healthsvc.Service
}

func NewHealthHandler(service *healthsvc.Service) *HealthHandler {
	return &HealthHandler{
		service: service,
	}
}

func (h *HealthHandler) Health(ctx context.Context, req *xov1.HealthRequest) (*xov1.HealthResponse, error) {
	if h.service.Ready(ctx) {
		return &xov1.HealthResponse{
			Status:  xov1.HealthStatus_HEALTH_STATUS_SERVING,
			Message: "ok",
		}, nil
	}

	return &xov1.HealthResponse{
		Status:  xov1.HealthStatus_HEALTH_STATUS_NOT_SERVING,
		Message: "not ready",
	}, nil
}
