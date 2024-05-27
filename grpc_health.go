package connect

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthCheckService handler for health check rpc
type HealthCheckService struct{}

// RegisterHealthCheckService init health check service
func RegisterHealthCheckService(registrar grpc.ServiceRegistrar, customHandler grpc_health_v1.HealthServer) {
	if customHandler != nil {
		grpc_health_v1.RegisterHealthServer(registrar, customHandler)
		return
	}

	// use default handler
	grpc_health_v1.RegisterHealthServer(registrar, HealthCheckService{})
}

// Check :nodoc:
func (s HealthCheckService) Check(_ context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

// Watch :nodoc:
func (s HealthCheckService) Watch(_ *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	return server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}
