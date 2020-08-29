package health

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func RegisterHealthServer(s *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(s, newHandler())
}

type healthImpl struct{}

func newHandler() grpc_health_v1.HealthServer {
	return &healthImpl{}
}

func (h *healthImpl) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *healthImpl) Watch(req *grpc_health_v1.HealthCheckRequest, w grpc_health_v1.Health_WatchServer) error {
	return nil
}
