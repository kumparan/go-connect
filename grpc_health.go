package connect

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// RegisterHealthCheckService init health check service
func RegisterHealthCheckService(appName string, registrar grpc.ServiceRegistrar, customHandler grpc_health_v1.HealthServer) {
	if customHandler != nil {
		grpc_health_v1.RegisterHealthServer(registrar, customHandler)
		return
	}

	defaultHealthServer := health.NewServer()                                                 // initialize health server
	defaultHealthServer.SetServingStatus(appName, grpc_health_v1.HealthCheckResponse_SERVING) // set default status

	// use default handler
	grpc_health_v1.RegisterHealthServer(registrar, defaultHealthServer)
}
