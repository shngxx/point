package health

// HealthCheck defines the interface for health checks
type HealthCheck interface {
	// Check performs a health check and returns an error if unhealthy
	Check() error
}

