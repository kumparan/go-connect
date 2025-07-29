package connect

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// RegisterHealthCheckService init health check service
func RegisterHealthCheckService(registrar grpc.ServiceRegistrar, customHandler grpc_health_v1.HealthServer) {
	if customHandler != nil {
		grpc_health_v1.RegisterHealthServer(registrar, customHandler)
		return
	}

	// use default health server
	grpc_health_v1.RegisterHealthServer(registrar, health.NewServer())
}
